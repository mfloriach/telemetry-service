package httpapi

import (
	"context"
	"errors"

	"server/internal/buffer"
)

// ErrInvalidLimit indicates the requested number of records is outside the allowed range.
var ErrInvalidLimit = errors.New("n must be an integer between 1 and 100")

// LatestSource provides the latest telemetry records.
type LatestSource interface {
	Latest(n int) []buffer.Record
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
	if n < 1 || n > 100 {
		return nil, ErrInvalidLimit
	}
	records := s.source.Latest(n)
	payloads := make(LatestResponse, 0, len(records))
	for _, rec := range records {
		payloads = append(payloads, rec.Payload)
	}
	return payloads, nil
}
