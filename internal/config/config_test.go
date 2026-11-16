package config

import (
	"io"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		env     map[string]string
		want    Config
		wantErr bool
	}{
		{
			name: "defaults with env ws url",
			env:  map[string]string{"TELEMETRY_WS_URL": "ws://example.test/telemetry"},
			want: Config{TelemetryWSURL: "ws://example.test/telemetry", HTTPAddr: ":8080"},
		},
		{
			name: "env http addr overrides default",
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"HTTP_ADDR":        "127.0.0.1:9000",
			},
			want: Config{TelemetryWSURL: "ws://example.test/telemetry", HTTPAddr: "127.0.0.1:9000"},
		},
		{
			name: "flag ws url overrides env",
			args: []string{"-telemetry-ws-url=ws://flag.test/ws"},
			env:  map[string]string{"TELEMETRY_WS_URL": "ws://env.test/ws"},
			want: Config{TelemetryWSURL: "ws://flag.test/ws", HTTPAddr: ":8080"},
		},
		{
			name: "flag http addr overrides env",
			args: []string{"-http-addr=:7777"},
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"HTTP_ADDR":        "127.0.0.1:9000",
			},
			want: Config{TelemetryWSURL: "ws://example.test/telemetry", HTTPAddr: ":7777"},
		},
		{
			name: "flags override both env values",
			args: []string{"-telemetry-ws-url=ws://flag.test/ws", "-http-addr=:7777"},
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://env.test/ws",
				"HTTP_ADDR":        "127.0.0.1:9000",
			},
			want: Config{TelemetryWSURL: "ws://flag.test/ws", HTTPAddr: ":7777"},
		},
		{
			name:    "missing required ws url",
			env:     map[string]string{},
			wantErr: true,
		},
		{
			name:    "missing required ws url when only http addr set",
			env:     map[string]string{"HTTP_ADDR": ":9000"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			getenv := func(key string) string {
				if tt.env == nil {
					return ""
				}
				return tt.env[key]
			}

			got, err := Load(tt.args, getenv, io.Discard)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Load() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Load() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Fatalf("Load() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
