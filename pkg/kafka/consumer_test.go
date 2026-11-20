package kafka

import (
	"strings"
	"testing"
	"time"

	segkafka "github.com/segmentio/kafka-go"
)

func TestNewConsumer_RequiredValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		brokers []string
		topic   string
		group   string
		wantErr string
	}{
		{name: "no brokers", brokers: nil, topic: "telemetry", group: "ingest", wantErr: "broker"},
		{name: "empty topic", brokers: []string{"localhost:9092"}, topic: "", group: "ingest", wantErr: "topic"},
		{name: "blank topic", brokers: []string{"localhost:9092"}, topic: "  ", group: "ingest", wantErr: "topic"},
		{name: "empty group", brokers: []string{"localhost:9092"}, topic: "telemetry", group: "", wantErr: "group"},
		{name: "blank group", brokers: []string{"localhost:9092"}, topic: "telemetry", group: " \t ", wantErr: "group"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			consumer, err := NewConsumer(tt.brokers, tt.topic, tt.group)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NewConsumer() error = %v, want containing %q", err, tt.wantErr)
			}
			if consumer != nil {
				t.Fatal("NewConsumer() returned a consumer for invalid input")
			}
		})
	}
}

func TestNewConsumer_Defaults(t *testing.T) {
	t.Parallel()

	consumer, err := NewConsumer([]string{"localhost:9092"}, "telemetry", "ingest")
	if err != nil {
		t.Fatalf("NewConsumer() unexpected error: %v", err)
	}
	t.Cleanup(func() {
		if err := consumer.Close(); err != nil {
			t.Errorf("Close() unexpected error: %v", err)
		}
	})

	cfg := consumer.reader.Config()
	if cfg.MinBytes != 1 {
		t.Fatalf("MinBytes = %d, want library default (1)", cfg.MinBytes)
	}
	if cfg.MaxBytes != 1_000_000 {
		t.Fatalf("MaxBytes = %d, want library default (1000000)", cfg.MaxBytes)
	}
	if cfg.MaxWait != 10*time.Second {
		t.Fatalf("MaxWait = %s, want library default (10s)", cfg.MaxWait)
	}
	if cfg.StartOffset != 0 {
		t.Fatalf("StartOffset = %d, want library default (0)", cfg.StartOffset)
	}
}

func TestNewConsumer_Options(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  []ConsumerOption
		check func(t *testing.T, cfg segkafka.ReaderConfig)
	}{
		{
			name: "fetch size",
			opts: []ConsumerOption{WithMinBytes(512), WithMaxBytes(4096)},
			check: func(t *testing.T, cfg segkafka.ReaderConfig) {
				t.Helper()
				if cfg.MinBytes != 512 {
					t.Fatalf("MinBytes = %d, want 512", cfg.MinBytes)
				}
				if cfg.MaxBytes != 4096 {
					t.Fatalf("MaxBytes = %d, want 4096", cfg.MaxBytes)
				}
			},
		},
		{
			name: "max wait",
			opts: []ConsumerOption{WithMaxWait(2 * time.Second)},
			check: func(t *testing.T, cfg segkafka.ReaderConfig) {
				t.Helper()
				if cfg.MaxWait != 2*time.Second {
					t.Fatalf("MaxWait = %s, want 2s", cfg.MaxWait)
				}
			},
		},
		{
			name: "start offset first",
			opts: []ConsumerOption{WithStartOffset(StartOffsetFirst)},
			check: func(t *testing.T, cfg segkafka.ReaderConfig) {
				t.Helper()
				if cfg.StartOffset != segkafka.FirstOffset {
					t.Fatalf("StartOffset = %d, want %d", cfg.StartOffset, segkafka.FirstOffset)
				}
			},
		},
		{
			name: "start offset last",
			opts: []ConsumerOption{WithStartOffset(StartOffsetLast)},
			check: func(t *testing.T, cfg segkafka.ReaderConfig) {
				t.Helper()
				if cfg.StartOffset != segkafka.LastOffset {
					t.Fatalf("StartOffset = %d, want %d", cfg.StartOffset, segkafka.LastOffset)
				}
			},
		},
		{
			name: "all options",
			opts: []ConsumerOption{
				WithMinBytes(256),
				WithMaxBytes(8192),
				WithMaxWait(500 * time.Millisecond),
				WithStartOffset(StartOffsetLast),
			},
			check: func(t *testing.T, cfg segkafka.ReaderConfig) {
				t.Helper()
				if cfg.MinBytes != 256 {
					t.Fatalf("MinBytes = %d, want 256", cfg.MinBytes)
				}
				if cfg.MaxBytes != 8192 {
					t.Fatalf("MaxBytes = %d, want 8192", cfg.MaxBytes)
				}
				if cfg.MaxWait != 500*time.Millisecond {
					t.Fatalf("MaxWait = %s, want 500ms", cfg.MaxWait)
				}
				if cfg.StartOffset != segkafka.LastOffset {
					t.Fatalf("StartOffset = %d, want %d", cfg.StartOffset, segkafka.LastOffset)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			consumer, err := NewConsumer([]string{"localhost:9092"}, "telemetry", "ingest", tt.opts...)
			if err != nil {
				t.Fatalf("NewConsumer() unexpected error: %v", err)
			}
			t.Cleanup(func() {
				if err := consumer.Close(); err != nil {
					t.Errorf("Close() unexpected error: %v", err)
				}
			})

			tt.check(t, consumer.reader.Config())
		})
	}
}

func TestNewConsumer_RejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []ConsumerOption
		wantErr string
	}{
		{name: "zero min bytes", opts: []ConsumerOption{WithMinBytes(0)}, wantErr: "min bytes"},
		{name: "negative min bytes", opts: []ConsumerOption{WithMinBytes(-1)}, wantErr: "min bytes"},
		{name: "zero max bytes", opts: []ConsumerOption{WithMaxBytes(0)}, wantErr: "max bytes"},
		{name: "negative max bytes", opts: []ConsumerOption{WithMaxBytes(-1)}, wantErr: "max bytes"},
		{name: "zero max wait", opts: []ConsumerOption{WithMaxWait(0)}, wantErr: "max wait"},
		{name: "negative max wait", opts: []ConsumerOption{WithMaxWait(-time.Second)}, wantErr: "max wait"},
		{name: "unsupported start offset", opts: []ConsumerOption{WithStartOffset(StartOffset(99))}, wantErr: "start offset"},
		{name: "min bytes above max bytes", opts: []ConsumerOption{WithMinBytes(4096), WithMaxBytes(1024)}, wantErr: "minimum batch size"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			consumer, err := NewConsumer([]string{"localhost:9092"}, "telemetry", "ingest", tt.opts...)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("NewConsumer() error = %v, want containing %q", err, tt.wantErr)
			}
			if consumer != nil {
				t.Fatal("NewConsumer() returned a consumer for invalid options")
			}
		})
	}
}

func TestNewConsumer_MinBytesWithoutMaxBytesReturnsError(t *testing.T) {
	t.Parallel()

	consumer, err := NewConsumer(
		[]string{"localhost:9092"},
		"telemetry",
		"ingest",
		WithMinBytes(512),
	)
	if err == nil {
		t.Fatal("NewConsumer() error = nil, want an error instead of a panic")
	}
	if consumer != nil {
		t.Fatal("NewConsumer() returned a consumer for an unrepresentable fetch size")
	}
}
