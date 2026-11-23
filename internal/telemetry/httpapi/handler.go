// Package httpapi provides the telemetry HTTP API server, handlers, and middleware.
package httpapi

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	httpServer "server/pkg/http"

	"github.com/go-playground/validator/v10"
)

const defaultLatestN = 10

type LatestResponse []json.RawMessage

type TelemetryService interface {
	Latest(ctx context.Context, n int) (LatestResponse, error)
}

type latestQuery struct {
	N int `validate:"min=1,max=100"`
}

type Handler struct {
	svc      TelemetryService
	logger   *slog.Logger
	validate *validator.Validate
}

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

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /telemetry/latest", h.handleLatest)
}

func (h *Handler) handleLatest(w http.ResponseWriter, r *http.Request) {
	query := latestQuery{N: defaultLatestN}

	if raw := r.URL.Query().Get("n"); raw != "" {
		n, ok := httpServer.ParseN(raw)
		if !ok {
			httpServer.WriteError(r.Context(), w, http.StatusBadRequest,
				"invalid query parameter: n must be an integer between 1 and 100", h.logger)
			return
		}
		query.N = n
	}

	if err := h.validate.Struct(query); err != nil {
		httpServer.WriteError(r.Context(), w, http.StatusBadRequest, "invalid query parameter: n must be an integer between 1 and 100", h.logger)
		return
	}

	payloads, err := h.svc.Latest(r.Context(), query.N)
	if err != nil {
		status := http.StatusInternalServerError
		msg := "internal server error"
		h.logger.ErrorContext(r.Context(), "failed to fetch latest telemetry", "error", err)
		httpServer.WriteError(r.Context(), w, status, msg, h.logger)
		return
	}

	httpServer.WriteJSON(r.Context(), w, http.StatusOK, payloads, h.logger)
}
