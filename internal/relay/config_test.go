package relay

import (
	"io"
	"testing"
)

func TestLoad_RequiresTelemetryWSURL(t *testing.T) {
	t.Parallel()

	_, err := Load(nil, func(string) string { return "" }, io.Discard)
	if err == nil {
		t.Fatal("expected error for missing TELEMETRY_WS_URL")
	}
	if got := err.Error(); got != "TELEMETRY_WS_URL is required (set env var or -telemetry-ws-url)" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_RequiresKafkaBrokers(t *testing.T) {
	t.Parallel()

	_, err := Load(nil, func(key string) string {
		if key == "TELEMETRY_WS_URL" {
			return "ws://example.test/telemetry"
		}
		return ""
	}, io.Discard)
	if err == nil {
		t.Fatal("expected error for missing KAFKA_BROKERS")
	}
	if got := err.Error(); got != "KAFKA_BROKERS is required (set env var or -kafka-brokers)" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLoad_DefaultsAndBrokers(t *testing.T) {
	t.Parallel()

	cfg, err := Load(nil, func(key string) string {
		switch key {
		case "TELEMETRY_WS_URL":
			return "ws://example.test/telemetry"
		case "KAFKA_BROKERS":
			return "localhost:9092, localhost:9093"
		default:
			return ""
		}
	}, io.Discard)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.KafkaTopic != "telemetry.raw" {
		t.Fatalf("expected default topic telemetry.raw, got %q", cfg.KafkaTopic)
	}
	brokers := cfg.Brokers()
	if len(brokers) != 2 || brokers[0] != "localhost:9092" || brokers[1] != "localhost:9093" {
		t.Fatalf("unexpected brokers: %#v", brokers)
	}
}
