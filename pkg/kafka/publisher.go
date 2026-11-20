// Package kafka provides Kafka-backed telemetry producers and consumers.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
)

// Publisher publishes telemetry records to a Kafka topic.
type Publisher struct {
	writer *kafka.Writer
}

// NewPublisher creates a Kafka publisher for the given brokers and topic.
func NewPublisher(brokers []string, topic string, opts ...PublisherOption) (*Publisher, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}
	if strings.TrimSpace(topic) == "" {
		return nil, errors.New("kafka topic is required")
	}

	cfg := publisherConfig{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, fmt.Errorf("configure kafka publisher: %w", err)
		}
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: cfg.requiredAcks,
		MaxAttempts:  cfg.maxAttempts,
		BatchSize:    cfg.batchSize,
		BatchTimeout: cfg.batchTimeout,
	}
	return &Publisher{writer: w}, nil
}

// Publish writes a single telemetry record.
func (p *Publisher) Publish(ctx context.Context, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Value: value})
}

// Close releases the underlying Kafka writer.
func (p *Publisher) Close() error {
	return p.writer.Close()
}
