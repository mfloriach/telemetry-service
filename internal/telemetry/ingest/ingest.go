package ingest

import (
	"context"
	"errors"
	"fmt"
)

type Consumer interface {
	Consume(ctx context.Context) ([]byte, error)
	Close() error
}

// Buffer stores ingested telemetry payloads.
type Buffer interface {
	Add(payload []byte)
}

// SessionIngester connects to an upstream Kafka consumer and appends received messages to a buffer
// for a single session.
type SessionIngester struct {
	consumer Consumer
	buffer   Buffer
}

// NewSessionIngester creates a session ingester.
func NewSessionIngester(consumer Consumer, buffer Buffer) *SessionIngester {
	return &SessionIngester{
		consumer: consumer,
		buffer:   buffer,
	}
}

// Run connects to a Kafka consumer and reads messages until ctx is done.
// It returns ctx.Err() when canceled, otherwise the consumer read error.
func (i *SessionIngester) Run(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		msg, err := i.consumer.Consume(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return ctx.Err()
			}
			return fmt.Errorf("read kafka consumer: %w", err)
		}
		i.buffer.Add(msg)
	}
}
