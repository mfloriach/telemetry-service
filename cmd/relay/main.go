// Package main runs the telemetry Kafka relay binary.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"server/internal/relay"
	"server/pkg/kafka"
	"server/pkg/logging"
	"server/pkg/ws"

	_ "github.com/joho/godotenv/autoload"
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

	publisher, err := kafka.NewPublisher(cfg.Brokers(), cfg.KafkaTopic)
	if err != nil {
		return fmt.Errorf("create kafka publisher: %w", err)
	}
	defer func() {
		if closeErr := publisher.Close(); closeErr != nil {
			logger.ErrorContext(context.Background(), "close kafka publisher", "error", closeErr)
		}
	}()

	handler, err := relay.NewHandler(publisher)
	if err != nil {
		return fmt.Errorf("create relay handler: %w", err)
	}

	srv, err := ws.NewServer(cfg.WSAddr, handler.Handle, ws.WithServerLogger(logger))
	if err != nil {
		return fmt.Errorf("create ws server: %w", err)
	}

	logger.InfoContext(ctx, "relay listening", "addr", srv.Addr(), "topic", cfg.KafkaTopic)

	if err := srv.Run(ctx); err != nil {
		return fmt.Errorf("ws server: %w", err)
	}

	return nil
}
