// Package relay wires device telemetry received over WebSocket to the Kafka publisher.
package relay

import (
	"context"
	"errors"
	"fmt"
)

// Publisher publishes telemetry records to Kafka.
type Publisher interface {
	Publish(ctx context.Context, value []byte) error
}

// Handler publishes telemetry messages received by the WebSocket server to Kafka.
type Handler struct {
	publisher Publisher
}

// NewHandler creates a handler that publishes through publisher.
func NewHandler(publisher Publisher) (*Handler, error) {
	if publisher == nil {
		return nil, errors.New("relay: publisher is required")
	}

	return &Handler{publisher: publisher}, nil
}

// Handle publishes payload to Kafka unchanged. A publish failure is returned so
// the server can close the originating device connection.
func (h *Handler) Handle(ctx context.Context, _ string, payload []byte) error {
	if err := h.publisher.Publish(ctx, payload); err != nil {
		return fmt.Errorf("relay: publish telemetry record: %w", err)
	}

	fmt.Println("received")

	return nil
}
