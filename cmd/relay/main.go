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

	publisher, err := kafka.NewKafkaPublisher(cfg.Brokers(), cfg.KafkaTopic)
	if err != nil {
		return fmt.Errorf("create kafka publisher: %w", err)
	}
	defer func() {
		if closeErr := publisher.Close(); closeErr != nil {
			logger.ErrorContext(context.Background(), "close kafka publisher", "error", closeErr)
		}
	}()

	return nil
}
