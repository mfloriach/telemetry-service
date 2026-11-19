package relay

import (
	"context"
	"errors"
	"fmt"
)

type Publisher interface {
	Publish(ctx context.Context, value []byte) error
	Close() error
}

type Conn interface {
	Read(ctx context.Context) (string, []byte, error)
	Close(code int, reason string) error
}

type runner struct {
	conn      Conn
	publisher Publisher
}

func NewRunner(conn Conn, publisher Publisher) *runner {
	return &runner{
		conn:      conn,
		publisher: publisher,
	}
}

func (r *runner) Run(ctx context.Context) error {
	defer func() {
		_ = r.conn.Close(1000, "relay stopping")
	}()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := r.runOnce(ctx)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ctx.Err()
		}
	}
}

func (r *runner) runOnce(ctx context.Context) error {
	for {
		_, msg, err := r.conn.Read(ctx)
		if err != nil {
			return err
		}

		if err := r.publisher.Publish(ctx, msg); err != nil {
			return fmt.Errorf("publish telemetry record to kafka: %w", err)
		}
	}
}
