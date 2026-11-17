package relay

import (
	"context"
	"errors"
	"fmt"

	"server/internal/utils/kafka"
	"server/internal/utils/ws"
)

// Runner consumes telemetry records from an upstream WebSocket and publishes them to Kafka.
type runner struct {
	dialer    ws.Dialer
	publisher kafka.Publisher
}

// NewRunner creates a relay runner.
func NewRunner(dialer ws.Dialer, publisher kafka.Publisher) *runner {
	return &runner{
		dialer:    dialer,
		publisher: publisher,
	}
}

// Run maintains the relay loop until ctx is canceled.
func (r *runner) Run(ctx context.Context, wsURL string) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := r.runOnce(ctx, wsURL)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ctx.Err()
		}
	}
}

func (r *runner) runOnce(ctx context.Context, wsURL string) error {
	conn, err := r.dialer.Dial(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("dial upstream websocket: %w", err)
	}
	defer func() {
		_ = conn.Close(1000, "relay stopping")
	}()

	for {
		msg, err := conn.Read(ctx)
		if err != nil {
			return err
		}

		if err := r.publisher.Publish(ctx, msg.Data); err != nil {
			return fmt.Errorf("publish telemetry record to kafka: %w", err)
		}
	}
}
