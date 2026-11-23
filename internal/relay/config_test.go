package relay

import (
	"io"
	"slices"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		env         map[string]string
		wantErr     string
		wantWSAddr  string
		wantTopic   string
		wantBrokers []string
	}{
		{
			name:        "default listen address and topic",
			env:         map[string]string{"KAFKA_BROKERS": "localhost:9092"},
			wantWSAddr:  ":9000",
			wantTopic:   "telemetry.raw",
			wantBrokers: []string{"localhost:9092"},
		},
		{
			name:        "listen address from environment",
			env:         map[string]string{"KAFKA_BROKERS": "localhost:9092", "WS_ADDR": "127.0.0.1:9100"},
			wantWSAddr:  "127.0.0.1:9100",
			wantTopic:   "telemetry.raw",
			wantBrokers: []string{"localhost:9092"},
		},
		{
			name:        "flag listen address wins over environment",
			args:        []string{"-ws-addr=127.0.0.1:9200"},
			env:         map[string]string{"KAFKA_BROKERS": "localhost:9092", "WS_ADDR": "127.0.0.1:9100"},
			wantWSAddr:  "127.0.0.1:9200",
			wantTopic:   "telemetry.raw",
			wantBrokers: []string{"localhost:9092"},
		},
		{
			name:        "brokers are split and trimmed",
			env:         map[string]string{"KAFKA_BROKERS": "localhost:9092, localhost:9093"},
			wantWSAddr:  ":9000",
			wantTopic:   "telemetry.raw",
			wantBrokers: []string{"localhost:9092", "localhost:9093"},
		},
		{
			name:    "malformed listen address",
			env:     map[string]string{"KAFKA_BROKERS": "localhost:9092", "WS_ADDR": "not-an-address"},
			wantErr: "invalid WS_ADDR",
		},
		{
			name:    "listen address without port",
			env:     map[string]string{"KAFKA_BROKERS": "localhost:9092", "WS_ADDR": "localhost"},
			wantErr: "invalid WS_ADDR",
		},
		{
			name:    "missing brokers",
			env:     map[string]string{"WS_ADDR": "127.0.0.1:9100"},
			wantErr: "KAFKA_BROKERS is required",
		},
		{
			name:        "stale telemetry websocket url is ignored",
			env:         map[string]string{"KAFKA_BROKERS": "localhost:9092", "TELEMETRY_WS_URL": "ws://stale.test/telemetry"},
			wantWSAddr:  ":9000",
			wantTopic:   "telemetry.raw",
			wantBrokers: []string{"localhost:9092"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			getenv := func(key string) string {
				return tt.env[key]
			}
			cfg, err := Load(tt.args, getenv, io.Discard)

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Load() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() unexpected error: %v", err)
			}
			if cfg.WSAddr != tt.wantWSAddr {
				t.Fatalf("WSAddr = %q, want %q", cfg.WSAddr, tt.wantWSAddr)
			}
			if cfg.KafkaTopic != tt.wantTopic {
				t.Fatalf("KafkaTopic = %q, want %q", cfg.KafkaTopic, tt.wantTopic)
			}
			if got := cfg.Brokers(); !slices.Equal(got, tt.wantBrokers) {
				t.Fatalf("Brokers() = %#v, want %#v", got, tt.wantBrokers)
			}
		})
	}
}
