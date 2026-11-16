// Package logging constructs configured slog loggers with context-aware request correlation.
package logging

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"server/internal/config"
)

// RequestIDFromContext extracts a request/correlation ID from context.
type RequestIDFromContext func(context.Context) string

// NewLogger builds a configured slog logger and wraps it to include request IDs from context.
func NewLogger(cfg config.Config, w io.Writer, requestID RequestIDFromContext) *slog.Logger {
	level := parseLevel(cfg.LogLevel)
	opts := &slog.HandlerOptions{Level: level}

	var base slog.Handler
	switch strings.ToLower(cfg.LogFormat) {
	case "json":
		base = slog.NewJSONHandler(w, opts)
	default:
		base = slog.NewTextHandler(w, opts)
	}

	return slog.New(&contextHandler{next: base, requestID: requestID})
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type contextHandler struct {
	next      slog.Handler
	requestID RequestIDFromContext
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.requestID != nil {
		if id := h.requestID(ctx); id != "" {
			r = r.Clone()
			r.AddAttrs(slog.String("request_id", id))
		}
	}
	return h.next.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{next: h.next.WithAttrs(attrs), requestID: h.requestID}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{next: h.next.WithGroup(name), requestID: h.requestID}
}
