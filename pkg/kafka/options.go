package kafka

import (
	"fmt"
	"time"

	segkafka "github.com/segmentio/kafka-go"
)

// Acks selects how many broker replicas must acknowledge a write before it is
// considered successful.
type Acks int

const (
	// AcksNone does not wait for broker acknowledgement.
	AcksNone Acks = iota
	// AcksLeader waits for the topic leader to acknowledge the write.
	AcksLeader
	// AcksAll waits for every in-sync replica to acknowledge the write.
	AcksAll
)

func (a Acks) toKafka() (segkafka.RequiredAcks, error) {
	switch a {
	case AcksNone:
		return segkafka.RequireNone, nil
	case AcksLeader:
		return segkafka.RequireOne, nil
	case AcksAll:
		return segkafka.RequireAll, nil
	default:
		return 0, fmt.Errorf("unsupported required acknowledgements: %d", int(a))
	}
}

// StartOffset selects where a consumer group begins when no offset is
// committed for a partition.
type StartOffset int

const (
	// StartOffsetDefault leaves the choice to the Kafka client default.
	StartOffsetDefault StartOffset = iota
	// StartOffsetFirst starts at the oldest available record.
	StartOffsetFirst
	// StartOffsetLast starts at the newest available record.
	StartOffsetLast
)

func (o StartOffset) toKafka() (int64, error) {
	switch o {
	case StartOffsetDefault:
		return 0, nil
	case StartOffsetFirst:
		return segkafka.FirstOffset, nil
	case StartOffsetLast:
		return segkafka.LastOffset, nil
	default:
		return 0, fmt.Errorf("unsupported start offset: %d", int(o))
	}
}

// PublisherOption configures a Publisher created by NewPublisher.
type PublisherOption func(*publisherConfig) error

type publisherConfig struct {
	requiredAcks segkafka.RequiredAcks
	maxAttempts  int
	batchSize    int
	batchTimeout time.Duration
}

// WithRequiredAcks sets how many broker replicas must acknowledge a write.
func WithRequiredAcks(acks Acks) PublisherOption {
	return func(cfg *publisherConfig) error {
		converted, err := acks.toKafka()
		if err != nil {
			return err
		}
		cfg.requiredAcks = converted
		return nil
	}
}

// WithMaxAttempts sets the maximum number of delivery attempts per message.
func WithMaxAttempts(attempts int) PublisherOption {
	return func(cfg *publisherConfig) error {
		if err := requirePositive("max attempts", attempts); err != nil {
			return err
		}
		cfg.maxAttempts = attempts
		return nil
	}
}

// WithBatchSize sets the maximum number of messages batched per write.
func WithBatchSize(size int) PublisherOption {
	return func(cfg *publisherConfig) error {
		if err := requirePositive("batch size", size); err != nil {
			return err
		}
		cfg.batchSize = size
		return nil
	}
}

// WithBatchTimeout sets how long the writer waits before flushing a batch.
func WithBatchTimeout(timeout time.Duration) PublisherOption {
	return func(cfg *publisherConfig) error {
		if err := requirePositiveDuration("batch timeout", timeout); err != nil {
			return err
		}
		cfg.batchTimeout = timeout
		return nil
	}
}

// ConsumerOption configures a Consumer created by NewConsumer.
type ConsumerOption func(*consumerConfig) error

type consumerConfig struct {
	minBytes    int
	maxBytes    int
	maxWait     time.Duration
	startOffset int64
}

// WithMinBytes sets the minimum number of bytes the consumer fetches.
func WithMinBytes(bytes int) ConsumerOption {
	return func(cfg *consumerConfig) error {
		if err := requirePositive("min bytes", bytes); err != nil {
			return err
		}
		cfg.minBytes = bytes
		return nil
	}
}

// WithMaxBytes sets the maximum number of bytes the consumer fetches.
func WithMaxBytes(bytes int) ConsumerOption {
	return func(cfg *consumerConfig) error {
		if err := requirePositive("max bytes", bytes); err != nil {
			return err
		}
		cfg.maxBytes = bytes
		return nil
	}
}

// WithMaxWait sets how long the consumer waits for enough data to fill a fetch.
func WithMaxWait(wait time.Duration) ConsumerOption {
	return func(cfg *consumerConfig) error {
		if err := requirePositiveDuration("max wait", wait); err != nil {
			return err
		}
		cfg.maxWait = wait
		return nil
	}
}

// WithStartOffset sets where a partition without a committed offset starts.
func WithStartOffset(offset StartOffset) ConsumerOption {
	return func(cfg *consumerConfig) error {
		converted, err := offset.toKafka()
		if err != nil {
			return err
		}
		cfg.startOffset = converted
		return nil
	}
}

func requirePositive(name string, value int) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive, got %d", name, value)
	}
	return nil
}

func requirePositiveDuration(name string, value time.Duration) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive, got %s", name, value)
	}
	return nil
}
