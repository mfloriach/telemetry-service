package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

const defaultHTTPAddr = ":8080"

// Config holds runtime configuration for the telemetry server.
type Config struct {
	TelemetryWSURL string
	HTTPAddr       string
}

// Load reads configuration from environment variables and command-line flags.
// Flags take precedence over environment variables. HTTPAddr defaults to ":8080".
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
	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	wsURL := firstNonEmpty(*wsURLFlag, getenv("TELEMETRY_WS_URL"))
	httpAddr := firstNonEmpty(*httpAddrFlag, getenv("HTTP_ADDR"), defaultHTTPAddr)

	if wsURL == "" {
		return Config{}, errors.New("TELEMETRY_WS_URL is required (set env var or -telemetry-ws-url)")
	}

	return Config{
		TelemetryWSURL: wsURL,
		HTTPAddr:       httpAddr,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
