// Package relay wires device telemetry received over WebSocket to the Kafka publisher.
package relay

import (
	"context"
	"fmt"
)

type Publisher interface {
	Publish(ctx context.Context, value []byte) error
}

type Handler struct {
	publisher Publisher
}

func NewHandler(publisher Publisher) (*Handler, error) {
	return &Handler{publisher: publisher}, nil
}

func (h *Handler) Handle(ctx context.Context, _ string, payload []byte) error {
	if err := h.publisher.Publish(ctx, payload); err != nil {
		return fmt.Errorf("relay: publish telemetry record: %w", err)
	}

	return nil
}
