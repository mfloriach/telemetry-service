package kafka

import (
	"context"
	"errors"
	"strings"

	"github.com/segmentio/kafka-go"
)

type Publisher interface {
	Publish(ctx context.Context, value []byte) error
	Close() error
}

type kafkaPublisher struct {
	writer *kafka.Writer
}

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

func (p *kafkaPublisher) Publish(ctx context.Context, value []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{Value: value})
}

func (p *kafkaPublisher) Close() error {
	return p.writer.Close()
}
