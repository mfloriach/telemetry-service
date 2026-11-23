# ADR 0006: Kafka telemetry relay as a separate process

- Status: Accepted (amended by [ADR-0009](0009-relay-hosts-websocket-ingest-server.md))
- Date: 2026-11-16
- Deciders: Project maintainer (approved in change discussion)

## Context

The system ingests upstream device telemetry in `cmd/telemetry-server`, buffers recent records in memory, and serves them over HTTP. There was no mechanism to forward ingested telemetry into an event-streaming system for downstream consumers. The change required adding Kafka publishing and local Kafka infrastructure without modifying existing telemetry-server runtime behavior.

## Decision

Introduce a standalone relay binary (`cmd/relay`) that hosts the device WebSocket ingest endpoint and publishes each received JSON telemetry record to a configured Kafka topic. Provide local development Kafka via `docker-compose.yml` using a single-broker KRaft configuration. The ingest endpoint and device-facing transport replaced the original upstream-dialer role; see ADR-0009.

- **Runtime/config:** relay requires `KAFKA_BROKERS`, uses `KAFKA_TOPIC` (default `telemetry.raw`), hosts the device WebSocket endpoint on `WS_ADDR` (default `:9000`, fixed path `/telemetry`), and supports optional logging environment settings consistent with project conventions. Relay no longer reads `TELEMETRY_WS_URL` (amended by ADR-0009).
- **Behavior:** publish original telemetry record content as Kafka message values; accept concurrent device connections and close only the affected connection when a publish ultimately fails. Devices are responsible for reconnecting and resending (amended by ADR-0009).
- **Dependency:** use `github.com/segmentio/kafka-go` as the Kafka client library.
- **Boundaries:** do not change existing `cmd/telemetry-server`, `internal/*` behavior, or tests. Minimal manifest/documentation/example updates (`go.mod`, `go.sum`, `.env.example`, README) are allowed only as needed for relay/Compose usability.

## Consequences

### Positive

- Enables downstream event consumption via Kafka without coupling to or modifying the existing telemetry-server process.
- Keeps failure domains separate and preserves ADR 0004's single-process ingest+serve design for the existing server.
- Provides a reproducible local Kafka environment for development and validation.

### Negative

- Adds an external dependency and operational components (Kafka broker, additional process).
- The relay is the single WebSocket ingest point, so devices cannot deliver telemetry while it is unavailable.
- Publish semantics are best-effort/at-least-once across reconnects or producer retries unless strengthened later.

## Alternatives Considered

### Alternative 1: Publish from `cmd/telemetry-server` ingest pipeline
- **Pros**: single upstream connection and integrated flow.
- **Cons**: requires modifying existing application behavior/code paths.
- **Why not**: conflicts with the change constraint to keep existing runtime behavior unchanged.

### Alternative 2: Poll `GET /telemetry/latest` and forward to Kafka
- **Pros**: no direct upstream connection in relay.
- **Cons**: coupling to HTTP API, polling latency/gaps/duplicates, weaker streaming semantics.
- **Why not**: less reliable and less direct than relaying source telemetry records.

### Alternative 3: Provide Kafka/Compose but no separate relay process
- **Pros**: smaller runtime surface.
- **Cons**: does not deliver publish capability.
- **Why not**: fails the core requirement to publish telemetry events to Kafka.

## References

- `openspec/changes/add-kafka-relay/proposal.md`
- `openspec/changes/add-kafka-relay/design.md`
- `openspec/changes/add-kafka-relay/specs/telemetry-kafka-relay/spec.md`
- `cmd/relay/main.go`
- `internal/relay/config.go`
- `internal/relay/handler.go`
- `docker-compose.yml`
- `docs/adr/0009-relay-hosts-websocket-ingest-server.md`
- `docs/adr/0004-resilient-single-process-ingest-and-serve.md`
