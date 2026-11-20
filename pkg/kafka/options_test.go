package kafka

import (
	"strings"
	"testing"
	"time"

	segkafka "github.com/segmentio/kafka-go"
)

func TestAcksToKafka(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		acks    Acks
		want    segkafka.RequiredAcks
		wantErr string
	}{
		{name: "none", acks: AcksNone, want: segkafka.RequireNone},
		{name: "leader", acks: AcksLeader, want: segkafka.RequireOne},
		{name: "all", acks: AcksAll, want: segkafka.RequireAll},
		{name: "unsupported", acks: Acks(99), wantErr: "unsupported required acknowledgements"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.acks.toKafka()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("toKafka() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("toKafka() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("toKafka() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStartOffsetToKafka(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		offset  StartOffset
		want    int64
		wantErr string
	}{
		{name: "default", offset: StartOffsetDefault, want: 0},
		{name: "first", offset: StartOffsetFirst, want: segkafka.FirstOffset},
		{name: "last", offset: StartOffsetLast, want: segkafka.LastOffset},
		{name: "unsupported", offset: StartOffset(99), wantErr: "unsupported start offset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.offset.toKafka()
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("toKafka() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("toKafka() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("toKafka() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRequirePositive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr string
	}{
		{name: "positive", value: 1},
		{name: "zero", value: 0, wantErr: "max attempts must be positive"},
		{name: "negative", value: -1, wantErr: "max attempts must be positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := requirePositive("max attempts", tt.value)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("requirePositive() unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("requirePositive() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestRequirePositiveDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   time.Duration
		wantErr string
	}{
		{name: "positive", value: time.Second},
		{name: "zero", value: 0, wantErr: "batch timeout must be positive"},
		{name: "negative", value: -time.Second, wantErr: "batch timeout must be positive"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := requirePositiveDuration("batch timeout", tt.value)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("requirePositiveDuration() unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("requirePositiveDuration() error = %v, want containing %q", err, tt.wantErr)
			}
		})
	}
}
