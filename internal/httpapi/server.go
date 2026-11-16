package httpapi

import (
	"log/slog"
	"net/http"
	"time"
)

// NewMux creates an HTTP mux with telemetry routes registered.
func NewMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

// NewServer creates an HTTP server for the telemetry API on the given address.
// Middleware is registered centrally around the mux.
func NewServer(addr string, h *Handler, logger *slog.Logger) *http.Server {
	if logger == nil {
		logger = slog.Default()
	}
	limiter := NewRateLimiter(120, time.Minute)
	handler := Chain(NewMux(h),
		RecoveryMiddleware(logger),
		RequestIDMiddleware(),
		LoggingMiddleware(logger),
		BodyLimitMiddleware(1<<20),
		RateLimitMiddleware(limiter),
		RequestTimeoutMiddleware(5*time.Second),
	)
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
