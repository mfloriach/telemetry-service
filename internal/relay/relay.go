// Package relay wires upstream telemetry reads to the Kafka publisher.
package relay

import (
	"context"
	"errors"
	"fmt"
)

// Publisher publishes telemetry records to Kafka.
type Publisher interface {
	Publish(ctx context.Context, value []byte) error
	Close() error
}

// Conn reads telemetry records from the upstream connection.
type Conn interface {
	Read(ctx context.Context) (string, []byte, error)
	Close(code int, reason string) error
}

// Runner reads telemetry records and publishes them to Kafka.
type Runner struct {
	conn      Conn
	publisher Publisher
}

// NewRunner creates a runner that reads from conn and publishes to publisher.
func NewRunner(conn Conn, publisher Publisher) *Runner {
	return &Runner{
		conn:      conn,
		publisher: publisher,
	}
}

// Run reads telemetry records until ctx is done or a read or publish fails.
func (r *Runner) Run(ctx context.Context) error {
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

func (r *Runner) runOnce(ctx context.Context) error {
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
