// Package kafka provides Kafka-backed telemetry producers and consumers.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"

	segkafka "github.com/segmentio/kafka-go"
)

// Consumer reads telemetry records from a Kafka topic.
type Consumer struct {
	reader *segkafka.Reader
}

// NewConsumer creates a Kafka consumer for the given brokers, topic, and group.
func NewConsumer(brokers []string, topic string, groupID string, opts ...ConsumerOption) (*Consumer, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}
	if strings.TrimSpace(topic) == "" {
		return nil, errors.New("kafka topic is required")
	}
	if strings.TrimSpace(groupID) == "" {
		return nil, errors.New("kafka consumer group is required")
	}

	cfg := consumerConfig{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			return nil, fmt.Errorf("configure kafka consumer: %w", err)
		}
	}

	readerCfg := segkafka.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		MinBytes:    cfg.minBytes,
		MaxBytes:    cfg.maxBytes,
		MaxWait:     cfg.maxWait,
		StartOffset: cfg.startOffset,
	}
	if err := readerCfg.Validate(); err != nil {
		return nil, fmt.Errorf("configure kafka consumer: %w", err)
	}

	return &Consumer{reader: segkafka.NewReader(readerCfg)}, nil
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
