// Package config loads and validates telemetry server runtime configuration.
package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	defaultHTTPAddr = ":8080"
	defaultAppEnv   = "development"
	defaultLogLevel = "info"
)

// Config holds runtime configuration for the telemetry server.
type Config struct {
	TelemetryWSURL string
	HTTPAddr       string
	AppEnv         string
	LogLevel       string
	LogFormat      string
}

// Load reads configuration from environment variables and command-line flags.
// Flags take precedence over environment variables.
// HTTPAddr defaults to ":8080". AppEnv defaults to "development".
// LogLevel defaults to "info". LogFormat defaults from AppEnv unless explicitly set.
func Load(args []string, getenv func(string) string, out io.Writer) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	if out == nil {
		out = io.Discard
	}

	fs := flag.NewFlagSet("telemetry-server", flag.ContinueOnError)
	fs.SetOutput(out)

	wsURLFlag := fs.String("telemetry-ws-url", "", "WebSocket URL for upstream telemetry source")
	httpAddrFlag := fs.String("http-addr", "", "HTTP listen address (default :8080)")
	appEnvFlag := fs.String("app-env", "", "Application environment (default development)")
	logLevelFlag := fs.String("log-level", "", "Log level: debug, info, warn, error (default info)")
	logFormatFlag := fs.String("log-format", "", "Log format: json or text (default derived from APP_ENV)")
	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	wsURL := firstNonEmpty(*wsURLFlag, getenv("TELEMETRY_WS_URL"))
	httpAddr := firstNonEmpty(*httpAddrFlag, getenv("HTTP_ADDR"), defaultHTTPAddr)
	appEnv := strings.ToLower(strings.TrimSpace(firstNonEmpty(*appEnvFlag, getenv("APP_ENV"), defaultAppEnv)))

	logLevel, err := normalizeLogLevel(firstNonEmpty(*logLevelFlag, getenv("LOG_LEVEL"), defaultLogLevel))
	if err != nil {
		return Config{}, err
	}

	logFormat, err := resolveLogFormat(firstNonEmpty(*logFormatFlag, getenv("LOG_FORMAT")), appEnv)
	if err != nil {
		return Config{}, err
	}

	if wsURL == "" {
		return Config{}, errors.New("TELEMETRY_WS_URL is required (set env var or -telemetry-ws-url)")
	}

	return Config{
		TelemetryWSURL: wsURL,
		HTTPAddr:       httpAddr,
		AppEnv:         appEnv,
		LogLevel:       logLevel,
		LogFormat:      logFormat,
	}, nil
}

func normalizeLogLevel(v string) (string, error) {
	level := strings.ToLower(strings.TrimSpace(v))
	switch level {
	case "debug", "info", "warn", "error":
		return level, nil
	default:
		return "", fmt.Errorf("invalid LOG_LEVEL %q: must be one of debug, info, warn, error", v)
	}
}

func resolveLogFormat(explicit, appEnv string) (string, error) {
	format := strings.ToLower(strings.TrimSpace(explicit))
	if format == "" {
		if appEnv == "production" {
			return "json", nil
		}
		return "text", nil
	}
	switch format {
	case "json", "text":
		return format, nil
	default:
		return "", fmt.Errorf("invalid LOG_FORMAT %q: must be one of json, text", explicit)
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
