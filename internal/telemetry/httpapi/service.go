package httpapi

import (
	"context"
	"encoding/json"
)

// LatestSource provides the latest telemetry payloads.
type LatestSource interface {
	Latest(n int) []json.RawMessage
}

// Service implements telemetry business logic for HTTP endpoints.
type Service struct {
	source LatestSource
}

// NewService creates a telemetry service.
func NewService(source LatestSource) *Service {
	return &Service{source: source}
}

// Latest returns up to n latest telemetry payloads sorted by ts descending.
func (s *Service) Latest(_ context.Context, n int) (LatestResponse, error) {
	payloads := s.source.Latest(n)
	if payloads == nil {
		payloads = []json.RawMessage{}
	}
	return LatestResponse(payloads), nil
}
