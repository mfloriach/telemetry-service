package httpapi

import (
	"context"
	"encoding/json"
)

type Record struct {
	Payload json.RawMessage
	TS      int64
}

// LatestSource provides the latest telemetry records.
type LatestSource interface {
	Latest(n int) []Record
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
	records := s.source.Latest(n)
	payloads := make(LatestResponse, 0, len(records))
	for _, rec := range records {
		payloads = append(payloads, rec.Payload)
	}
	return payloads, nil
}
