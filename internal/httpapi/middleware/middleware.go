package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

// RequestIDFromContext returns the request ID stored by RequestIDMiddleware.
func RequestIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(requestIDContextKey).(string)
	return v
}

// Middleware wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order so the first listed is the outermost.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(errorResponse{Error: msg}); err != nil {
		slog.Default().Error("failed to encode json response", "error", err)
	}
}
