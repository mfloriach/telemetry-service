package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"server/internal/config"
)

func TestNewLogger_FormatSelectionAndLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		cfg       config.Config
		log       func(*slog.Logger)
		wantJSON  bool
		wantDebug bool
	}{
		{
			name: "json format info level suppresses debug",
			cfg: config.Config{
				LogLevel:  "info",
				LogFormat: "json",
			},
			log: func(l *slog.Logger) {
				l.Debug("debug-msg")
				l.Info("info-msg")
			},
			wantJSON:  true,
			wantDebug: false,
		},
		{
			name: "text format debug level emits debug",
			cfg: config.Config{
				LogLevel:  "debug",
				LogFormat: "text",
			},
			log: func(l *slog.Logger) {
				l.Debug("debug-msg")
			},
			wantJSON:  false,
			wantDebug: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			logger := NewLogger(tt.cfg.LogFormat, tt.cfg.LogLevel, &buf, nil)
			tt.log(logger)
			out := buf.String()
			if strings.TrimSpace(out) == "" {
				t.Fatalf("expected log output")
			}
			first := strings.TrimSpace(strings.SplitN(out, "\n", 2)[0])
			isJSON := strings.HasPrefix(first, "{")
			if isJSON != tt.wantJSON {
				t.Fatalf("json output = %v, want %v; output=%q", isJSON, tt.wantJSON, out)
			}
			hasDebug := strings.Contains(out, "debug-msg")
			if hasDebug != tt.wantDebug {
				t.Fatalf("debug present = %v, want %v; output=%q", hasDebug, tt.wantDebug, out)
			}
		})
	}
}

func TestNewLogger_RequestIDFromContext(t *testing.T) {
	t.Parallel()

	type ctxKey string
	const key ctxKey = "request_id"

	var buf bytes.Buffer
	cfg := config.Config{LogLevel: "info", LogFormat: "json"}
	logger := NewLogger(cfg.LogFormat, cfg.LogLevel, &buf, func(ctx context.Context) string {
		v, _ := ctx.Value(key).(string)
		return v
	})

	ctx := context.WithValue(context.Background(), key, "req-123")
	logger.InfoContext(ctx, "hello", "k", "v")

	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatalf("expected log output")
	}

	var rec map[string]any
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		t.Fatalf("unmarshal json log: %v; line=%q", err, line)
	}
	if got, _ := rec["request_id"].(string); got != "req-123" {
		t.Fatalf("request_id = %q, want %q; record=%v", got, "req-123", rec)
	}
	if got, _ := rec["msg"].(string); got != "hello" {
		t.Fatalf("msg = %q, want %q; record=%v", got, "hello", rec)
	}
}

func TestNewLogger_WithoutRequestID(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	cfg := config.Config{LogLevel: "info", LogFormat: "json"}
	logger := NewLogger(cfg.LogFormat, cfg.LogLevel, &buf, func(context.Context) string { return "" })
	logger.InfoContext(context.Background(), "hello")

	line := strings.TrimSpace(buf.String())
	var rec map[string]any
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		t.Fatalf("unmarshal json log: %v; line=%q", err, line)
	}
	if _, ok := rec["request_id"]; ok {
		t.Fatalf("unexpected request_id in record: %v", rec)
	}
}
