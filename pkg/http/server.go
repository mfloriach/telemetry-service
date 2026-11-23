package http

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON[T any](ctx context.Context, w http.ResponseWriter, status int, v T, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.ErrorContext(ctx, "failed to encode json response", "error", err)
	}
}

func WriteError(ctx context.Context, w http.ResponseWriter, status int, msg string, logger *slog.Logger) {
	WriteJSON(ctx, w, status, ErrorResponse{Error: msg}, logger)
}

func ParseN(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return n, true
}
