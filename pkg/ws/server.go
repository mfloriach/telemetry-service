package ws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	// telemetryPath is the fixed HTTP path where the server accepts WebSocket handshakes.
	telemetryPath = "/telemetry"

	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 5 * time.Second
)

// MessageHandler processes one inbound telemetry message delivered by the server.
// Returning an error closes the originating connection with a server-error status.
type MessageHandler func(ctx context.Context, messageType string, payload []byte) error

// ServerOption configures a Server built by NewServer.
type ServerOption func(*serverConfig)

type serverConfig struct {
	readLimit int64
	logger    *slog.Logger
}

func defaultServerConfig() serverConfig {
	return serverConfig{readLimit: DefaultReadLimit, logger: slog.Default()}
}

// WithServerReadLimit sets the largest inbound message the server accepts, in bytes.
// Values must be positive.
func WithServerReadLimit(limit int64) ServerOption {
	return func(cfg *serverConfig) {
		cfg.readLimit = limit
	}
}

// WithServerLogger sets the logger used to report connection activity. A nil
// logger leaves the default slog logger in place.
func WithServerLogger(logger *slog.Logger) ServerOption {
	return func(cfg *serverConfig) {
		if logger != nil {
			cfg.logger = logger
		}
	}
}

// Server accepts device WebSocket connections on a fixed telemetry path and
// delivers each inbound message to a MessageHandler.
type Server struct {
	addr      string
	handler   MessageHandler
	readLimit int64
	logger    *slog.Logger

	mu        sync.Mutex
	conns     map[*websocket.Conn]struct{}
	closing   bool
	boundAddr string
}

// NewServer creates a server that listens on addr and handles messages with handler.
// The address and handler are required; the read limit must be positive.
func NewServer(addr string, handler MessageHandler, opts ...ServerOption) (*Server, error) {
	if strings.TrimSpace(addr) == "" {
		return nil, errors.New("ws: listen address is required")
	}
	if handler == nil {
		return nil, errors.New("ws: message handler is required")
	}

	cfg := defaultServerConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.readLimit <= 0 {
		return nil, fmt.Errorf("ws: read limit must be positive, got %d", cfg.readLimit)
	}

	return &Server{
		addr:      addr,
		handler:   handler,
		readLimit: cfg.readLimit,
		logger:    cfg.logger,
		conns:     make(map[*websocket.Conn]struct{}),
	}, nil
}

// Addr returns the address the server is listening on. Before Run binds the
// listener it returns the configured address.
func (s *Server) Addr() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.boundAddr != "" {
		return s.boundAddr
	}
	return s.addr
}

// Run binds the listener and serves device connections until ctx is canceled or
// the server fails. A requested shutdown closes live connections with a
// going-away status and returns nil.
func (s *Server) Run(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("ws: listen on %s: %w", s.addr, err)
	}

	s.mu.Lock()
	s.boundAddr = listener.Addr().String()
	s.mu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+telemetryPath, s.handleConnection)
	httpServer := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	serveErr := make(chan error, 1)
	go func() {
		err := httpServer.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErr <- err
	}()

	select {
	case err := <-serveErr:
		return err
	case <-ctx.Done():
	}

	s.beginShutdown()
	s.closeAll(websocket.StatusGoingAway, "server shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		_ = httpServer.Close()
	}
	if err := <-serveErr; err != nil {
		return err
	}

	return nil
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(s.readLimit)

	if !s.track(conn) {
		_ = conn.Close(websocket.StatusGoingAway, "server shutting down")
		return
	}
	defer s.untrack(conn)

	ctx := r.Context()
	remote := r.RemoteAddr
	s.logger.InfoContext(ctx, "websocket connection opened", "remote_addr", remote)
	defer s.logger.InfoContext(ctx, "websocket connection closed", "remote_addr", remote)

	for {
		messageType, payload, err := conn.Read(ctx)
		if err != nil {
			s.handleReadError(conn, err)
			return
		}

		s.logger.InfoContext(ctx, "websocket message received",
			"remote_addr", remote,
			"type", messageTypeName(messageType),
			"bytes", len(payload),
		)

		if err := s.handler(ctx, messageTypeName(messageType), payload); err != nil {
			s.logger.ErrorContext(ctx, "websocket message handler failed",
				"remote_addr", remote,
				"error", err,
			)
			_ = conn.Close(websocket.StatusInternalError, "failed to process telemetry")
			return
		}
	}
}

func (s *Server) handleReadError(conn *websocket.Conn, err error) {
	status := websocket.CloseStatus(err)
	if status == websocket.StatusNormalClosure {
		return
	}
	if s.isClosing() {
		return
	}

	code := status
	switch {
	case errors.Is(err, websocket.ErrMessageTooBig):
		code = websocket.StatusMessageTooBig
	case !closableStatus(code):
		code = websocket.StatusInternalError
	}

	if closeErr := conn.Close(code, "read failed"); closeErr != nil {
		return
	}
}

func (s *Server) track(conn *websocket.Conn) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closing {
		return false
	}
	s.conns[conn] = struct{}{}
	return true
}

func (s *Server) untrack(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.conns, conn)
}

func (s *Server) beginShutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closing = true
}

func (s *Server) isClosing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.closing
}

func (s *Server) closeAll(code websocket.StatusCode, reason string) {
	s.mu.Lock()
	conns := make([]*websocket.Conn, 0, len(s.conns))
	for conn := range s.conns {
		conns = append(conns, conn)
	}
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, conn := range conns {
		wg.Add(1)
		go func(conn *websocket.Conn) {
			defer wg.Done()
			_ = conn.Close(code, reason)
		}(conn)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(shutdownTimeout):
		for _, conn := range conns {
			_ = conn.CloseNow()
		}
	}
}

func messageTypeName(messageType websocket.MessageType) string {
	switch messageType {
	case websocket.MessageText:
		return "text"
	case websocket.MessageBinary:
		return "binary"
	default:
		return messageType.String()
	}
}

func closableStatus(code websocket.StatusCode) bool {
	switch code {
	case websocket.StatusNormalClosure,
		websocket.StatusGoingAway,
		websocket.StatusProtocolError,
		websocket.StatusUnsupportedData,
		websocket.StatusInvalidFramePayloadData,
		websocket.StatusPolicyViolation,
		websocket.StatusMessageTooBig,
		websocket.StatusMandatoryExtension,
		websocket.StatusInternalError:
		return true
	}

	return code >= 3000 && code <= 4999
}
