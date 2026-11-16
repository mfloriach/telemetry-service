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

	"server/internal/buffer"
	"server/internal/config"
	"server/internal/httpapi"
	"server/internal/ingest"
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

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	buf := buffer.NewRingBuffer(buffer.DefaultCapacity)
	ing := ingest.NewIngester(ingest.CoderDialer{}, buf, logger)
	handler := httpapi.NewHandler(buf)
	srv := httpapi.NewServer(cfg.HTTPAddr, handler)

	ingestDone := make(chan struct{})
	go func() {
		defer close(ingestDone)
		if err := ing.Run(ctx, cfg.TelemetryWSURL); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("ingestion stopped unexpectedly", "error", err)
		}
	}()

	httpErrCh := make(chan error, 1)
	go func() {
		logger.Info("http server listening", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			httpErrCh <- err
		}
		close(httpErrCh)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-httpErrCh:
		if err != nil {
			stop()
			return fmt.Errorf("http server: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown error", "error", err)
	}

	stop()
	<-ingestDone
	return nil
}
