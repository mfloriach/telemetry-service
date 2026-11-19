// Package kafka provides Kafka-backed telemetry producers and consumers.
package kafka

import (
	"context"
	"errors"
	"strings"

	"github.com/segmentio/kafka-go"
)

// Publisher publishes telemetry records to a Kafka topic.
type Publisher struct {
	writer *kafka.Writer
}

// NewPublisher creates a Kafka publisher for the given brokers and topic.
func NewPublisher(brokers []string, topic string) (*Publisher, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}
	if strings.TrimSpace(topic) == "" {
		return nil, errors.New("kafka topic is required")
	}
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
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
