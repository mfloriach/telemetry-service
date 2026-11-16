package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"server/internal/buffer"
)

// LatestSource provides the latest telemetry records for HTTP responses.
type LatestSource interface {
	Latest(n int) []buffer.Record
}

// Handler serves telemetry HTTP endpoints.
type Handler struct {
	source LatestSource
}

// NewHandler creates a telemetry HTTP handler.
func NewHandler(source LatestSource) *Handler {
	return &Handler{source: source}
}

// RegisterRoutes registers telemetry routes on mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /telemetry/latest", h.handleLatest)
}

// NewMux creates an HTTP mux with telemetry routes registered.
func NewMux(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return mux
}

// NewServer creates an HTTP server for the telemetry API on the given address.
func NewServer(addr string, h *Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           NewMux(h),
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func (h *Handler) handleLatest(w http.ResponseWriter, r *http.Request) {
	n, ok := parseN(r.URL.Query().Get("n"))
	if !ok {
		http.Error(w, "invalid query parameter: n must be an integer between 1 and 100", http.StatusBadRequest)
		return
	}

	records := h.source.Latest(n)
	payloads := make([]json.RawMessage, 0, len(records))
	for _, rec := range records {
		payloads = append(payloads, rec.Payload)
	}

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	if err := enc.Encode(payloads); err != nil {
		// Best-effort: headers/status may already be written.
		return
	}
}

func parseN(raw string) (int, bool) {
	if raw == "" {
		return 0, false
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false
	}
	if n < 1 || n > 100 {
		return 0, false
	}
	return n, true
}
