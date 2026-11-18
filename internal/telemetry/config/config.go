// Package config loads and validates telemetry server runtime configuration.
package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/ilyakaznacheev/cleanenv"
)

const (
	defaultHTTPAddr   = ":8080"
	defaultAppEnv     = "development"
	defaultLogLevel   = "info"
	defaultKafkaTopic = "telemetry.raw"
)

// Config holds runtime configuration for the telemetry server.
type Config struct {
	KafkaBrokers   []string `env:"KAFKA_BROKERS" validate:"required,min=1,dive,required"`
	KafkaTopic     string   `env:"KAFKA_TOPIC" env-default:"telemetry.raw" validate:"required"`
	HTTPAddr       string   `env:"HTTP_ADDR" env-default:":8080" validate:"required"`
	AppEnv         string   `env:"APP_ENV" env-default:"development" validate:"required"`
	LogLevel       string   `env:"LOG_LEVEL" env-default:"info" validate:"required,oneof=debug info warn error"`
	LogFormat      string   `env:"LOG_FORMAT" validate:"required,oneof=json text"`
}

// Load reads configuration from environment variables and command-line flags.
// Flags take precedence over environment variables.
// HTTPAddr defaults to ":8080". AppEnv defaults to "development".
// LogLevel defaults to "info". LogFormat defaults from AppEnv unless explicitly set.
func Load(args []string, out io.Writer) (Config, error) {
	if out == nil {
		out = io.Discard
	}

	var cfg Config

	// Load environment variables and defaults.
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("load environment: %w", err)
	}

	fs := flag.NewFlagSet("telemetry-server", flag.ContinueOnError)
	fs.SetOutput(out)

	brokers := fs.String("kafka-brokers", strings.Join(cfg.KafkaBrokers, ","), "Kafka bootstrap brokers")
	topic := fs.String("kafka-topic", cfg.KafkaTopic, "Kafka topic")
	httpAddr := fs.String("http-addr", cfg.HTTPAddr, "HTTP listen address")
	appEnv := fs.String("app-env", cfg.AppEnv, "Application environment")
	logLevel := fs.String("log-level", cfg.LogLevel, "Log level")
	logFormat := fs.String("log-format", cfg.LogFormat, "Log format")

	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	// Flags override environment variables.
	cfg.KafkaBrokers = splitCSV(*brokers)
	cfg.KafkaTopic = strings.TrimSpace(*topic)
	cfg.HTTPAddr = strings.TrimSpace(*httpAddr)
	cfg.AppEnv = strings.ToLower(strings.TrimSpace(*appEnv))
	cfg.LogLevel = strings.ToLower(strings.TrimSpace(*logLevel))
	cfg.LogFormat = strings.ToLower(strings.TrimSpace(*logFormat))

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
		var validationErrors validator.ValidationErrors
		if !errors.As(err, &validationErrors) {
			return fmt.Errorf("validate config: %w", err)
		}

		fieldNames := map[string]string{
			"KafkaBrokers":   "KAFKA_BROKERS",
			"KafkaTopic":     "KAFKA_TOPIC",
			"HTTPAddr":       "HTTP_ADDR",
			"AppEnv":         "APP_ENV",
			"LogLevel":       "LOG_LEVEL",
			"LogFormat":      "LOG_FORMAT",
		}

		fe := validationErrors[0]
		return fmt.Errorf("%s: validation failed (%s)", fieldNames[fe.Field()], fe.Tag())
	}

	return nil
}

func defaultLogFormat(appEnv string) string {
	if appEnv == "production" {
		return "json"
	}
	return "text"
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}
