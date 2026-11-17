// Package main runs the telemetry Kafka relay binary.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"server/internal/relay"
	"server/internal/utils/kafka"
	"server/internal/utils/logging"
	"server/internal/utils/ws"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	brokers := cfg.Brokers()
	if len(brokers) == 0 {
		return errors.New("KAFKA_BROKERS is required (set env var or -kafka-brokers)")
	}

	publisher, err := kafka.NewKafkaPublisher(brokers, cfg.KafkaTopic)
	if err != nil {
		return fmt.Errorf("create kafka publisher: %w", err)
	}
	defer func() {
		if closeErr := publisher.Close(); closeErr != nil {
			logger.ErrorContext(context.Background(), "close kafka publisher", "error", closeErr)
		}
	}()

	runner := relay.NewRunner(ws.CoderDialer{}, publisher)
	logger.InfoContext(ctx, "telemetry relay starting", "url", cfg.TelemetryWSURL, "topic", cfg.KafkaTopic, "brokers", brokers)

	if err := runner.Run(ctx, cfg.TelemetryWSURL); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("relay: %w", err)
	}

	logger.InfoContext(ctx, "telemetry relay stopped")
	return nil
}
