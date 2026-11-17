# Proposal

## Why

The system receives telemetry data over WebSocket connections but lacks a mechanism to forward those records to Kafka for downstream processing. Without this relay, real-time telemetry is not persisted or made available to Kafka consumers, limiting observability and analytics. This change adds a WebSocket-to-Kafka relay capability so telemetry data can be streamed asynchronously to downstream systems.

## What Changes

- **New capability**: `telemetry/relay` — relay telemetry records from WebSocket connections to Kafka
- **New file**: `internal/relay/relay.go` — Runner that consumes telemetry from WebSocket and publishes to Kafka
- **New file**: `internal/ingest/ingest.go` — SessionIngester that connects to WebSocket and buffers messages
- **Existing file**: `internal/utils/ws/ws.go` — Dialer and Conn interfaces for WebSocket connections
- **Existing file**: `internal/utils/ws/coder_dialer.go` — CoderDialer implementation for WebSocket dialing
- **Existing file**: `internal/utils/kafka/publisher.go` — Publisher interface and Kafka implementation
- **New test**: `internal/relay/relay_integration_test.go` — integration tests for WebSocket-to-Kafka flow
- **New test**: `internal/ingest/ingest_integration_test.go` — integration tests for buffer behavior

## Capabilities

### New Capabilities

- `telemetry/relay`: Relay telemetry records from WebSocket connections to Kafka for downstream processing and analytics

### Modified Capabilities

*(No existing capabilities modified)*

## Impact

- **New code**: `internal/relay/relay.go`, `internal/ingest/ingest.go`
- **New interfaces**: `internal/utils/ws/dialer.go` (Dialer), `internal/utils/kafka/publisher.go` (Publisher)
- **Dependency**: Kafka cluster configuration required for message production
- **Dependency**: WebSocket upstream endpoint configuration
- **Tests**: Integration tests in `internal/relay/` and `internal/ingest/`