// Package kafka provides Kafka-backed telemetry producers and consumers.
package kafka

import (
	"context"

	segkafka "github.com/segmentio/kafka-go"
)

// Consumer reads telemetry records from a Kafka topic.
type Consumer struct {
	reader *segkafka.Reader
}

// NewConsumer creates a Kafka consumer for the given brokers, topic, and group.
func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	r := segkafka.NewReader(segkafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &Consumer{reader: r}
}

// Consume reads the next telemetry record.
func (a *Consumer) Consume(ctx context.Context) ([]byte, error) {
	msg, err := a.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	return msg.Value, nil
}

// Close releases the underlying Kafka reader.
func (a *Consumer) Close() error {
	return a.reader.Close()
}
