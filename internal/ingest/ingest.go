package ingest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"server/internal/utils"
)

// Buffer stores ingested telemetry payloads.
type Buffer interface {
	Add(payload []byte)
}

// SessionIngester connects to an upstream WebSocket and appends received messages to a buffer
// for a single session.
type SessionIngester struct {
	dialer Dialer
	buffer Buffer
	logger *slog.Logger
}

// NewSessionIngester creates a session ingester.
func NewSessionIngester(dialer Dialer, buffer Buffer, logger *slog.Logger) *SessionIngester {
	if logger == nil {
		logger = slog.Default()
	}
	return &SessionIngester{
		dialer: dialer,
		buffer: buffer,
		logger: logger,
	}
}

// RunOnce connects to url and reads messages until ctx is done or the connection fails.
// It returns ctx.Err() when canceled, otherwise the connection/read error.
func (i *SessionIngester) RunOnce(ctx context.Context, url string) error {
	conn, err := i.dialer.Dial(ctx, url)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}
	i.logger.InfoContext(ctx, "telemetry websocket connected", "url", url)
	defer func() {
		_ = conn.Close(1000, "closing")
	}()

	for {
		msg, err := conn.Read(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return ctx.Err()
			}
			return fmt.Errorf("read websocket: %w", err)
		}
		i.buffer.Add(msg.Data)
	}
}

// RunWithRetry runs the session ingester with exponential backoff + jitter retry.
// It uses the utils retry mechanism and returns when ctx is canceled.
func (i *SessionIngester) RunWithRetry(ctx context.Context, url string) error {
	return utils.WithRetry(ctx, func(ctx context.Context) error {
		return i.RunOnce(ctx, url)
	}, 100*time.Millisecond, 5*time.Second, func(d time.Duration) time.Duration {
		if d <= 0 {
			return 0
		}
		half := d / 2
		if half <= 0 {
			return d
		}
		return half + rand.N(d-half+1) //nolint:gosec // non-cryptographic jitter for reconnect backoff
	}, time.Second)
}