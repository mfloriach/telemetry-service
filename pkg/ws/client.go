// Package ws provides WebSocket client and server primitives for moving telemetry
// between devices, the relay, and upstream sources.
package ws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

const (
	// DefaultHandshakeTimeout bounds the opening handshake when no timeout is configured.
	DefaultHandshakeTimeout = 10 * time.Second
	// DefaultReadLimit is the largest inbound message accepted when no limit is configured.
	DefaultReadLimit int64 = 1 << 20
)

// WebSocket is a connection to the upstream telemetry source.
type WebSocket struct {
	conn   *websocket.Conn
	logger *slog.Logger
	ctx    context.Context
}

// Option configures a WebSocket built by NewWebsocket.
type Option func(*config)

type config struct {
	handshakeTimeout     time.Duration
	readLimit            int64
	compression          bool
	compressionThreshold int
	subprotocols         []string
	header               http.Header
	logger               *slog.Logger
}

func defaultConfig() config {
	return config{
		handshakeTimeout: DefaultHandshakeTimeout,
		readLimit:        DefaultReadLimit,
	}
}

// WithHandshakeTimeout bounds the opening handshake. Values must be positive.
func WithHandshakeTimeout(timeout time.Duration) Option {
	return func(cfg *config) {
		cfg.handshakeTimeout = timeout
	}
}

// WithReadLimit sets the largest inbound message accepted, in bytes. Values must be positive.
func WithReadLimit(limit int64) Option {
	return func(cfg *config) {
		cfg.readLimit = limit
	}
}

// WithCompression enables or disables WebSocket compression and sets the minimum message
// size, in bytes, before compression is applied. A threshold is only valid while enabled.
func WithCompression(enabled bool, threshold int) Option {
	return func(cfg *config) {
		cfg.compression = enabled
		cfg.compressionThreshold = threshold
	}
}

// WithSubprotocols offers the given subprotocols during the opening handshake.
func WithSubprotocols(subprotocols ...string) Option {
	return func(cfg *config) {
		cfg.subprotocols = append([]string(nil), subprotocols...)
	}
}

// WithHeader adds a request header to the opening handshake.
func WithHeader(key, value string) Option {
	return func(cfg *config) {
		if cfg.header == nil {
			cfg.header = make(http.Header)
		}
		cfg.header.Set(key, value)
	}
}

// NewWebsocket dials the upstream WebSocket endpoint and returns a ready connection.
// The endpoint is required; every other setting is optional.
func NewWebsocket(ctx context.Context, url string, opts ...Option) (*WebSocket, error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if err := cfg.validate(url); err != nil {
		return nil, err
	}

	dialCtx, cancel := context.WithTimeout(ctx, cfg.handshakeTimeout)
	defer cancel()

	dialOpts := &websocket.DialOptions{
		HTTPHeader:   cfg.header,
		Subprotocols: cfg.subprotocols,
	}
	if cfg.compression {
		dialOpts.CompressionMode = websocket.CompressionContextTakeover
		dialOpts.CompressionThreshold = cfg.compressionThreshold
	}

	conn, _, err := websocket.Dial(dialCtx, url, dialOpts)
	if err != nil {
		return nil, fmt.Errorf("ws: dial %s: %w", url, err)
	}
	conn.SetReadLimit(cfg.readLimit)

	return &WebSocket{conn: conn, logger: cfg.logger, ctx: ctx}, nil
}

func (cfg config) validate(url string) error {
	switch {
	case url == "":
		return errors.New("ws: url is required")
	case cfg.handshakeTimeout <= 0:
		return fmt.Errorf("ws: handshake timeout must be positive, got %s", cfg.handshakeTimeout)
	case cfg.readLimit <= 0:
		return fmt.Errorf("ws: read limit must be positive, got %d", cfg.readLimit)
	case cfg.compressionThreshold != 0 && !cfg.compression:
		return fmt.Errorf("ws: compression threshold %d requires compression to be enabled", cfg.compressionThreshold)
	}
	return nil
}

// Read returns the next inbound message and its frame type.
func (c *WebSocket) Read(ctx context.Context) (string, []byte, error) {
	messageType, msg, err := c.conn.Read(ctx)
	if err != nil {
		return "", nil, err
	}
	return messageType.String(), msg, nil
}

// Close sends the given close code and reason and releases the connection.
func (c *WebSocket) Close(code int, reason string) error {
	return c.conn.Close(websocket.StatusCode(code), reason)
}
