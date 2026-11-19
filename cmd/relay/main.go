// Package main runs the telemetry Kafka relay binary.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"server/internal/relay"
	"server/pkg/kafka"
	"server/pkg/logging"

	"server/pkg/ws"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := relay.Load(os.Args[1:], os.Getenv, os.Stderr)
	if err != nil {
		return err
	}

	logger := logging.NewLogger(cfg.LogFormat, cfg.LogLevel, os.Stdout, nil)
	slog.SetDefault(logger)

	ctx := context.Background()

	publisher, err := kafka.NewPublisher(cfg.Brokers(), cfg.KafkaTopic)
	if err != nil {
		return fmt.Errorf("create kafka publisher: %w", err)
	}
	defer func() {
		if closeErr := publisher.Close(); closeErr != nil {
			logger.ErrorContext(ctx, "close kafka publisher", "error", closeErr)
		}
	}()

	conn, err := ws.NewWebsocket(ctx, cfg.TelemetryWSURL)
	if err != nil {
		return fmt.Errorf("ws connection fail: %w", err)
	}

	if err := relay.NewRunner(conn, publisher).Run(ctx); err != nil {
		return fmt.Errorf("ws connection fail: %w", err)
	}

	return nil
}
