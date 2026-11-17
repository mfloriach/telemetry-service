# Proposal

## Why

The `SessionIngester` in `internal/ingest/ingest.go` currently reads telemetry records from an upstream WebSocket connection. This creates a tight coupling between the ingest component and WebSocket-specific transport concerns (dialing, connection management, frame handling). Changing the ingest component to read from a Kafka consumer instead decouples ingest from WebSocket transport, aligns it with the existing Kafka-based relay architecture, and enables the system to consume telemetry records already persisted in Kafka topics.

## What Changes

- **Modified file**: `internal/ingest/ingest.go` — `SessionIngester` reads from a Kafka consumer instead of a WebSocket dialer
- **New interface**: `internal/utils/kafka/consumer.go` — `Consumer` interface for consuming messages from Kafka topics
- **Modified struct fields**: `SessionIngester` replaces `dialer ws.Dialer` with `consumer kafka.Consumer`
- **Modified constructor**: `NewSessionIngester` accepts a `kafka.Consumer` instead of `ws.Dialer`
- **Modified `Run` method**: reads messages from the Kafka consumer loop instead of WebSocket `conn.Read(ctx)`
- **BREAKING**: `SessionIngester` no longer accepts a WebSocket dialer or URL; callers must provide a Kafka consumer

## Capabilities

### New Capabilities

- `ingest/kafka-consumer`: Ingest telemetry records from a Kafka consumer into a buffer for downstream processing

### Modified Capabilities

- `telemetry-kafka-relay`: The relay's ingest source changes from WebSocket to Kafka consumer; requirements around upstream WebSocket connection no longer apply to this component

## Impact

- **Modified code**: `internal/ingest/ingest.go` — SessionIngester struct, constructor, and Run method
- **New code**: `internal/utils/kafka/consumer.go` — Consumer interface definition
- **Breaking change**: Callers of `NewSessionIngester` must pass a `kafka.Consumer` instead of `ws.Dialer`; the `url` parameter is removed from `Run`
- **Dependency**: Kafka cluster configuration required for consumer group setup
- **No impact on**: `internal/relay/relay.go` (unchanged), `internal/utils/ws` (unchanged)