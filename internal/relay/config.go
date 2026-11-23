// Package relay wires upstream telemetry reads to the Kafka publisher.
package relay

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	defaultAppEnv   = "development"
	defaultLogLevel = "info"
	defaultTopic    = "telemetry.raw"
	defaultWSAddr   = ":9000"
)

// Config holds runtime configuration for the telemetry Kafka relay.
type Config struct {
	WSAddr       string `validate:"required"`
	KafkaBrokers string `validate:"required"`
	KafkaTopic   string `validate:"required"`
	AppEnv       string `validate:"required"`
	LogLevel     string `validate:"required,oneof=debug info warn error"`
	LogFormat    string `validate:"required,oneof=json text"`
}

// Load reads configuration from environment variables and command-line flags.
// Flags take precedence over environment variables.
func Load(args []string, getenv func(string) string, out io.Writer) (Config, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	if out == nil {
		out = io.Discard
	}

	fs := flag.NewFlagSet("relay", flag.ContinueOnError)
	fs.SetOutput(out)

	wsAddrFlag := fs.String("ws-addr", "", "WebSocket listen address for device connections (default :9000)")
	brokersFlag := fs.String("kafka-brokers", "", "Comma-separated Kafka bootstrap broker addresses")
	topicFlag := fs.String("kafka-topic", "", "Kafka topic name (default telemetry.raw)")
	appEnvFlag := fs.String("app-env", "", "Application environment (default development)")
	logLevelFlag := fs.String("log-level", "", "Log level: debug, info, warn, error (default info)")
	logFormatFlag := fs.String("log-format", "", "Log format: json or text (default derived from APP_ENV)")
	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	cfg := Config{
		WSAddr:       strings.TrimSpace(firstNonEmpty(*wsAddrFlag, getenv("WS_ADDR"), defaultWSAddr)),
		KafkaBrokers: strings.TrimSpace(firstNonEmpty(*brokersFlag, getenv("KAFKA_BROKERS"))),
		KafkaTopic:   strings.TrimSpace(firstNonEmpty(*topicFlag, getenv("KAFKA_TOPIC"), defaultTopic)),
		AppEnv:       strings.ToLower(strings.TrimSpace(firstNonEmpty(*appEnvFlag, getenv("APP_ENV"), defaultAppEnv))),
		LogLevel:     strings.ToLower(strings.TrimSpace(firstNonEmpty(*logLevelFlag, getenv("LOG_LEVEL"), defaultLogLevel))),
		LogFormat:    strings.ToLower(strings.TrimSpace(firstNonEmpty(*logFormatFlag, getenv("LOG_FORMAT")))),
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
				case "WSAddr":
					return errors.New("WS_ADDR is required (set env var or -ws-addr)")
				case "KafkaBrokers":
					return errors.New("KAFKA_BROKERS is required (set env var or -kafka-brokers)")
				case "KafkaTopic":
					return errors.New("KAFKA_TOPIC is required (set env var or -kafka-topic)")
				case "LogLevel":
					return fmt.Errorf("invalid LOG_LEVEL %q: must be one of debug, info, warn, error", cfg.LogLevel)
				case "LogFormat":
					return fmt.Errorf("invalid LOG_FORMAT %q: must be one of json, text", cfg.LogFormat)
				}
			}
		}
		return fmt.Errorf("invalid config: %w", err)
	}

	if _, _, err := net.SplitHostPort(cfg.WSAddr); err != nil {
		return fmt.Errorf("invalid WS_ADDR %q: %w", cfg.WSAddr, err)
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

// Brokers returns configured broker addresses split by comma.
func (c Config) Brokers() []string {
	parts := strings.Split(c.KafkaBrokers, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
