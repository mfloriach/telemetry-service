package kafka

import (
	"context"
	"errors"
	"strings"

	"github.com/segmentio/kafka-go"
)

// Publisher publishes a telemetry record value to Kafka.
type Publisher interface {
	Publish(ctx context.Context, value []byte) error
	Close() error
}

// kafkaPublisher publishes messages to a Kafka topic using segmentio/kafka-go.
type kafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher creates a Kafka publisher for the given brokers and topic.
func NewKafkaPublisher(brokers []string, topic string) (*kafkaPublisher, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}
	if strings.TrimSpace(topic) == "" {
		return nil, errors.New("Kafka topic is required")
	}
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &kafkaPublisher{writer: w}, nil
}

// Publish writes a message with the provided value to the configured topic.
// It retries with exponential backoff + jitter on failure.
func (p *kafkaPublisher) Publish(ctx context.Context, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Value: value})
}

// Close closes the underlying Kafka writer.
func (p *kafkaPublisher) Close() error {
	return p.writer.Close()
}
