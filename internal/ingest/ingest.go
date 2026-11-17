package ingest

import (
	"context"
	"errors"
	"fmt"

	"server/internal/utils/ws"
)

// Buffer stores ingested telemetry payloads.
type Buffer interface {
	Add(payload []byte)
}

// SessionIngester connects to an upstream WebSocket and appends received messages to a buffer
// for a single session.
type SessionIngester struct {
	dialer ws.Dialer
	buffer Buffer
}

// NewSessionIngester creates a session ingester.
func NewSessionIngester(dialer ws.Dialer, buffer Buffer) *SessionIngester {
	return &SessionIngester{
		dialer: dialer,
		buffer: buffer,
	}
}

// RunOnce connects to url and reads messages until ctx is done or the connection fails.
// It returns ctx.Err() when canceled, otherwise the connection/read error.
func (i *SessionIngester) Run(ctx context.Context, url string) error {
	conn, err := i.dialer.Dial(ctx, url)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}
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
