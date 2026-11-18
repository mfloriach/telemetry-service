// Package httpapi provides the telemetry HTTP API server, handlers, and middleware.
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// LatestResponse is the JSON response body for GET /telemetry/latest.
type LatestResponse []json.RawMessage

// TelemetryService provides telemetry query operations for HTTP handlers.
type TelemetryService interface {
	Latest(ctx context.Context, n int) (LatestResponse, error)
}

type latestQuery struct {
	N int `validate:"required,min=1,max=100"`
}

// Handler serves telemetry HTTP endpoints.
type Handler struct {
	svc      TelemetryService
	logger   *slog.Logger
	validate *validator.Validate
}

// NewHandler creates a telemetry HTTP handler.
func NewHandler(svc TelemetryService, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	return &Handler{
		svc:      svc,
		logger:   logger,
		validate: validator.New(validator.WithRequiredStructEnabled()),
	}
}

// RegisterRoutes registers telemetry routes on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /telemetry/latest", h.handleLatest)
}

func (h *Handler) handleLatest(w http.ResponseWriter, r *http.Request) {
	n, ok := parseN(r.URL.Query().Get("n"))
	if !ok {
		writeError(r.Context(), w, http.StatusBadRequest, "invalid query parameter: n must be an integer between 1 and 100", h.logger)
		return
	}

	query := latestQuery{N: n}
	if err := h.validate.Struct(query); err != nil {
		writeError(r.Context(), w, http.StatusBadRequest, "invalid query parameter: n must be an integer between 1 and 100", h.logger)
		return
	}

	payloads, err := h.svc.Latest(r.Context(), query.N)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "internal server error"
		h.logger.ErrorContext(r.Context(), "failed to fetch latest telemetry", "error", err)
		writeError(r.Context(), w, status, msg, h.logger)
		return
	}

	writeJSON(r.Context(), w, http.StatusOK, payloads, h.logger)
}

func parseN(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	return n, true
}

// ErrorResponse is the JSON error response body.
type ErrorResponse struct {
	Error string `json:"error"`
}

func writeJSON[T any](ctx context.Context, w http.ResponseWriter, status int, v T, logger *slog.Logger) {
	if logger == nil {
		logger = slog.Default()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.ErrorContext(ctx, "failed to encode json response", "error", err)
	}
}

func writeError(ctx context.Context, w http.ResponseWriter, status int, msg string, logger *slog.Logger) {
	writeJSON(ctx, w, status, ErrorResponse{Error: msg}, logger)
}
