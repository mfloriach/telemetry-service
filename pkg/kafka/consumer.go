package kafka

import (
	"context"

	segkafka "github.com/segmentio/kafka-go"
)

type kafkaConsumerAdapter struct {
	reader *segkafka.Reader
}

func NewKafkaConsumer(brokers []string, topic string, groupID string) *kafkaConsumerAdapter {
	r := segkafka.NewReader(segkafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &kafkaConsumerAdapter{reader: r}
}

func (a *kafkaConsumerAdapter) Consume(ctx context.Context) ([]byte, error) {
	msg, err := a.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	return msg.Value, nil
}

func (a *kafkaConsumerAdapter) Close() error {
	return a.reader.Close()
}
