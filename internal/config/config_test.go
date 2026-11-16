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
			want: Config{
				TelemetryWSURL: "ws://example.test/telemetry",
				HTTPAddr:       ":8080",
				AppEnv:         "development",
				LogLevel:       "info",
				LogFormat:      "text",
			},
		},
		{
			name: "env http addr overrides default",
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"HTTP_ADDR":        "127.0.0.1:9000",
			},
			want: Config{
				TelemetryWSURL: "ws://example.test/telemetry",
				HTTPAddr:       "127.0.0.1:9000",
				AppEnv:         "development",
				LogLevel:       "info",
				LogFormat:      "text",
			},
		},
		{
			name: "flag ws url overrides env",
			args: []string{"-telemetry-ws-url=ws://flag.test/ws"},
			env:  map[string]string{"TELEMETRY_WS_URL": "ws://env.test/ws"},
			want: Config{
				TelemetryWSURL: "ws://flag.test/ws",
				HTTPAddr:       ":8080",
				AppEnv:         "development",
				LogLevel:       "info",
				LogFormat:      "text",
			},
		},
		{
			name: "flag http addr overrides env",
			args: []string{"-http-addr=:7777"},
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"HTTP_ADDR":        "127.0.0.1:9000",
			},
			want: Config{
				TelemetryWSURL: "ws://example.test/telemetry",
				HTTPAddr:       ":7777",
				AppEnv:         "development",
				LogLevel:       "info",
				LogFormat:      "text",
			},
		},
		{
			name: "flags override both env values",
			args: []string{"-telemetry-ws-url=ws://flag.test/ws", "-http-addr=:7777"},
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://env.test/ws",
				"HTTP_ADDR":        "127.0.0.1:9000",
			},
			want: Config{
				TelemetryWSURL: "ws://flag.test/ws",
				HTTPAddr:       ":7777",
				AppEnv:         "development",
				LogLevel:       "info",
				LogFormat:      "text",
			},
		},
		{
			name: "production app env defaults log format to json",
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"APP_ENV":          "Production",
			},
			want: Config{
				TelemetryWSURL: "ws://example.test/telemetry",
				HTTPAddr:       ":8080",
				AppEnv:         "production",
				LogLevel:       "info",
				LogFormat:      "json",
			},
		},
		{
			name: "explicit log format overrides env default",
			args: []string{"-log-format=text"},
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"APP_ENV":          "production",
			},
			want: Config{
				TelemetryWSURL: "ws://example.test/telemetry",
				HTTPAddr:       ":8080",
				AppEnv:         "production",
				LogLevel:       "info",
				LogFormat:      "text",
			},
		},
		{
			name: "log level and format normalize case and flags override env",
			args: []string{"-log-level=ERROR", "-log-format=JSON"},
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"APP_ENV":          "PRODUCTION",
				"LOG_LEVEL":        "debug",
				"LOG_FORMAT":       "text",
			},
			want: Config{
				TelemetryWSURL: "ws://example.test/telemetry",
				HTTPAddr:       ":8080",
				AppEnv:         "production",
				LogLevel:       "error",
				LogFormat:      "json",
			},
		},
		{
			name: "env log settings applied when flags absent",
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"APP_ENV":          "staging",
				"LOG_LEVEL":        "WARN",
				"LOG_FORMAT":       "JSON",
			},
			want: Config{
				TelemetryWSURL: "ws://example.test/telemetry",
				HTTPAddr:       ":8080",
				AppEnv:         "staging",
				LogLevel:       "warn",
				LogFormat:      "json",
			},
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
		{
			name: "invalid log level fails fast",
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"LOG_LEVEL":        "verbose",
			},
			wantErr: true,
		},
		{
			name: "invalid log format fails fast",
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
				"LOG_FORMAT":       "yaml",
			},
			wantErr: true,
		},
		{
			name: "invalid log format flag fails fast",
			args: []string{"-log-format=pretty"},
			env: map[string]string{
				"TELEMETRY_WS_URL": "ws://example.test/telemetry",
			},
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
