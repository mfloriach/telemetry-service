// Package kafka provides Kafka-backed telemetry producers and consumers.
package kafka

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
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

	// Ensure topic exists.
	if err := ensureTopic(brokers[0], topic); err != nil {
		return nil, fmt.Errorf("ensure kafka topic: %w", err)
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

func ensureTopic(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("dial kafka broker: %w", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("get kafka controller: %w", err)
	}

	controllerConn, err := kafka.Dial(
		"tcp",
		net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)),
	)
	if err != nil {
		return fmt.Errorf("dial kafka controller: %w", err)
	}
	defer controllerConn.Close()

	return controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     3,
		ReplicationFactor: 1,
	})
}
