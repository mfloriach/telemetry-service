package kafka

import (
	"strings"
	"testing"
	"time"

	segkafka "github.com/segmentio/kafka-go"
)

func TestNewPublisher_RequiredValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		brokers []string
		topic   string
		wantErr string
	}{
		{name: "no brokers", brokers: nil, topic: "telemetry", wantErr: "broker"},
		{name: "empty topic", brokers: []string{"localhost:9092"}, topic: "", wantErr: "topic"},
		{name: "blank topic", brokers: []string{"localhost:9092"}, topic: "   ", wantErr: "topic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			publisher, err := NewPublisher(tt.brokers, tt.topic)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NewPublisher() error = %v, want containing %q", err, tt.wantErr)
			}
			if publisher != nil {
				t.Fatal("NewPublisher() returned a publisher for invalid input")
			}
		})
	}
}

func TestNewPublisher_Defaults(t *testing.T) {
	t.Parallel()

	publisher, err := NewPublisher([]string{"localhost:9092"}, "telemetry")
	if err != nil {
		t.Fatalf("NewPublisher() unexpected error: %v", err)
	}

	if _, ok := publisher.writer.Balancer.(*segkafka.LeastBytes); !ok {
		t.Fatalf("Balancer = %T, want *kafka.LeastBytes", publisher.writer.Balancer)
	}
	if publisher.writer.RequiredAcks != segkafka.RequireNone {
		t.Fatalf("RequiredAcks = %v, want %v", publisher.writer.RequiredAcks, segkafka.RequireNone)
	}
	if publisher.writer.MaxAttempts != 0 {
		t.Fatalf("MaxAttempts = %d, want library default (0)", publisher.writer.MaxAttempts)
	}
	if publisher.writer.BatchSize != 0 {
		t.Fatalf("BatchSize = %d, want library default (0)", publisher.writer.BatchSize)
	}
	if publisher.writer.BatchTimeout != 0 {
		t.Fatalf("BatchTimeout = %s, want library default (0)", publisher.writer.BatchTimeout)
	}
}

func TestNewPublisher_Options(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  []PublisherOption
		check func(t *testing.T, w *segkafka.Writer)
	}{
		{
			name: "required acks",
			opts: []PublisherOption{WithRequiredAcks(AcksAll)},
			check: func(t *testing.T, w *segkafka.Writer) {
				t.Helper()
				if w.RequiredAcks != segkafka.RequireAll {
					t.Fatalf("RequiredAcks = %v, want %v", w.RequiredAcks, segkafka.RequireAll)
				}
			},
		},
		{
			name: "max attempts",
			opts: []PublisherOption{WithMaxAttempts(5)},
			check: func(t *testing.T, w *segkafka.Writer) {
				t.Helper()
				if w.MaxAttempts != 5 {
					t.Fatalf("MaxAttempts = %d, want 5", w.MaxAttempts)
				}
			},
		},
		{
			name: "batch size",
			opts: []PublisherOption{WithBatchSize(250)},
			check: func(t *testing.T, w *segkafka.Writer) {
				t.Helper()
				if w.BatchSize != 250 {
					t.Fatalf("BatchSize = %d, want 250", w.BatchSize)
				}
			},
		},
		{
			name: "batch timeout",
			opts: []PublisherOption{WithBatchTimeout(2 * time.Second)},
			check: func(t *testing.T, w *segkafka.Writer) {
				t.Helper()
				if w.BatchTimeout != 2*time.Second {
					t.Fatalf("BatchTimeout = %s, want 2s", w.BatchTimeout)
				}
			},
		},
		{
			name: "all options",
			opts: []PublisherOption{
				WithRequiredAcks(AcksLeader),
				WithMaxAttempts(3),
				WithBatchSize(10),
				WithBatchTimeout(500 * time.Millisecond),
			},
			check: func(t *testing.T, w *segkafka.Writer) {
				t.Helper()
				if w.RequiredAcks != segkafka.RequireOne {
					t.Fatalf("RequiredAcks = %v, want %v", w.RequiredAcks, segkafka.RequireOne)
				}
				if w.MaxAttempts != 3 {
					t.Fatalf("MaxAttempts = %d, want 3", w.MaxAttempts)
				}
				if w.BatchSize != 10 {
					t.Fatalf("BatchSize = %d, want 10", w.BatchSize)
				}
				if w.BatchTimeout != 500*time.Millisecond {
					t.Fatalf("BatchTimeout = %s, want 500ms", w.BatchTimeout)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			publisher, err := NewPublisher([]string{"localhost:9092"}, "telemetry", tt.opts...)
			if err != nil {
				t.Fatalf("NewPublisher() unexpected error: %v", err)
			}
			tt.check(t, publisher.writer)
		})
	}
}

func TestNewPublisher_RejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []PublisherOption
		wantErr string
	}{
		{name: "zero max attempts", opts: []PublisherOption{WithMaxAttempts(0)}, wantErr: "max attempts"},
		{name: "negative max attempts", opts: []PublisherOption{WithMaxAttempts(-1)}, wantErr: "max attempts"},
		{name: "zero batch size", opts: []PublisherOption{WithBatchSize(0)}, wantErr: "batch size"},
		{name: "negative batch size", opts: []PublisherOption{WithBatchSize(-1)}, wantErr: "batch size"},
		{name: "zero batch timeout", opts: []PublisherOption{WithBatchTimeout(0)}, wantErr: "batch timeout"},
		{name: "negative batch timeout", opts: []PublisherOption{WithBatchTimeout(-time.Second)}, wantErr: "batch timeout"},
		{name: "unsupported acks", opts: []PublisherOption{WithRequiredAcks(Acks(99))}, wantErr: "required acknowledgements"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			publisher, err := NewPublisher([]string{"localhost:9092"}, "telemetry", tt.opts...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NewPublisher() error = %v, want containing %q", err, tt.wantErr)
			}
			if publisher != nil {
				t.Fatal("NewPublisher() returned a publisher for invalid options")
			}
		})
	}
}
