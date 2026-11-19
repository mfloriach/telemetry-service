// Package ws provides the WebSocket client used to read upstream telemetry.
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

type WebSocket struct {
	conn   *websocket.Conn
	logger *slog.Logger
	ctx    context.Context
}

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

func WithHandshakeTimeout(timeout time.Duration) Option {
	return func(cfg *config) {
		cfg.handshakeTimeout = timeout
	}
}

func WithReadLimit(limit int64) Option {
	return func(cfg *config) {
		cfg.readLimit = limit
	}
}

func WithCompression(enabled bool, threshold int) Option {
	return func(cfg *config) {
		cfg.compression = enabled
		cfg.compressionThreshold = threshold
	}
}

func WithSubprotocols(subprotocols ...string) Option {
	return func(cfg *config) {
		cfg.subprotocols = append([]string(nil), subprotocols...)
	}
}

func WithHeader(key, value string) Option {
	return func(cfg *config) {
		if cfg.header == nil {
			cfg.header = make(http.Header)
		}
		cfg.header.Set(key, value)
	}
}

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

func (c *WebSocket) Read(ctx context.Context) (string, []byte, error) {
	messageType, msg, err := c.conn.Read(ctx)
	if err != nil {
		return "", nil, err
	}
	return messageType.String(), msg, nil
}

func (c *WebSocket) Close(code int, reason string) error {
	if err := c.conn.Close(websocket.StatusCode(code), reason); err != nil {
		return fmt.Errorf("err")
	}
	return nil
}
