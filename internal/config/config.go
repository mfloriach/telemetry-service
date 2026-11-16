// Package config loads and validates telemetry server runtime configuration.
package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	defaultHTTPAddr = ":8080"
	defaultAppEnv   = "development"
	defaultLogLevel = "info"
)

// Config holds runtime configuration for the telemetry server.
type Config struct {
	TelemetryWSURL string `validate:"required"`
	HTTPAddr       string `validate:"required"`
	AppEnv         string `validate:"required"`
	LogLevel       string `validate:"required,oneof=debug info warn error"`
	LogFormat      string `validate:"required,oneof=json text"`
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

	cfg := Config{
		TelemetryWSURL: strings.TrimSpace(firstNonEmpty(*wsURLFlag, getenv("TELEMETRY_WS_URL"))),
		HTTPAddr:       strings.TrimSpace(firstNonEmpty(*httpAddrFlag, getenv("HTTP_ADDR"), defaultHTTPAddr)),
		AppEnv:         strings.ToLower(strings.TrimSpace(firstNonEmpty(*appEnvFlag, getenv("APP_ENV"), defaultAppEnv))),
		LogLevel:       strings.ToLower(strings.TrimSpace(firstNonEmpty(*logLevelFlag, getenv("LOG_LEVEL"), defaultLogLevel))),
		LogFormat:      strings.ToLower(strings.TrimSpace(firstNonEmpty(*logFormatFlag, getenv("LOG_FORMAT")))),
	}
	if cfg.LogFormat == "" {
		cfg.LogFormat = defaultLogFormat(cfg.AppEnv)
	}

	if err := validateConfig(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func validateConfig(cfg Config) error {
	v := validator.New(validator.WithRequiredStructEnabled())
	if err := v.Struct(cfg); err != nil {
		var verrs validator.ValidationErrors
		if errors.As(err, &verrs) {
			for _, fe := range verrs {
				switch fe.Field() {
				case "TelemetryWSURL":
					return errors.New("TELEMETRY_WS_URL is required (set env var or -telemetry-ws-url)")
				case "LogLevel":
					return fmt.Errorf("invalid LOG_LEVEL %q: must be one of debug, info, warn, error", cfg.LogLevel)
				case "LogFormat":
					return fmt.Errorf("invalid LOG_FORMAT %q: must be one of json, text", cfg.LogFormat)
				}
			}
		}
		return fmt.Errorf("invalid config: %w", err)
	}
	return nil
}

func defaultLogFormat(appEnv string) string {
	if appEnv == "production" {
		return "json"
	}
	return "text"
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
