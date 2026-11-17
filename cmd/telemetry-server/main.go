// Package main runs the telemetry server binary.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server/internal/config"
	"server/internal/httpapi"
	"server/internal/httpapi/middleware"
	"server/internal/ingest"
	"server/internal/utils/buffer"
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
	cfg, err := config.Load(os.Args[1:], os.Getenv, os.Stderr)
	if err != nil {
		return err
	}

	logger := logging.NewLogger(cfg.LogFormat, cfg.LogLevel, os.Stdout, middleware.RequestIDFromContext)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	buf := buffer.NewRingBuffer(buffer.DefaultCapacity)
	sessionIng := ingest.NewSessionIngester(ws.CoderDialer{}, buf)
	svc := httpapi.NewService(buf)
	handler := httpapi.NewHandler(svc, logger)
	srv := httpapi.NewServer(cfg.HTTPAddr, handler, logger)

	ingestDone := make(chan struct{})
	go func() {
		defer close(ingestDone)
		if err := sessionIng.Run(ctx, cfg.TelemetryWSURL); err != nil && !errors.Is(err, context.Canceled) {
			logger.ErrorContext(ctx, "ingestion stopped unexpectedly", "error", err)
		}
	}()

	httpErrCh := make(chan error, 1)
	go func() {
		logger.InfoContext(ctx, "http server listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			httpErrCh <- err
		}
		close(httpErrCh)
	}()

	select {
	case <-ctx.Done():
		logger.InfoContext(ctx, "shutdown signal received")
	case err := <-httpErrCh:
		if err != nil {
			stop()
			return fmt.Errorf("http server: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.ErrorContext(shutdownCtx, "http server shutdown error", "error", err)
	}

	stop()
	<-ingestDone
	return nil
}
