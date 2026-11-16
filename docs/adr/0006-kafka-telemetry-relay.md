# ADR 0006: Kafka telemetry relay as a separate process

- Status: Accepted
- Date: 2026-11-16
- Deciders: Project maintainer (approved in change discussion)

## Context

The system ingests upstream satellite telemetry in `cmd/telemetry-server`, buffers recent records in memory, and serves them over HTTP. There was no mechanism to forward ingested telemetry into an event-streaming system for downstream consumers. The change required adding Kafka publishing and local Kafka infrastructure without modifying existing telemetry-server runtime behavior.

## Decision

Introduce a standalone relay binary (`cmd/relay`) that independently connects to the configured upstream WebSocket telemetry source and publishes each received JSON telemetry record to a configured Kafka topic. Provide local development Kafka via `docker-compose.yml` using a single-broker KRaft configuration.

- **Runtime/config:** relay requires `TELEMETRY_WS_URL` and `KAFKA_BROKERS`, uses `KAFKA_TOPIC` (default `telemetry.raw`), and supports optional logging environment settings consistent with project conventions.
- **Behavior:** publish original telemetry record content as Kafka message values; apply bounded reconnect/backoff handling for upstream instability.
- **Dependency:** use `github.com/segmentio/kafka-go` as the Kafka client library.
- **Boundaries:** do not change existing `cmd/telemetry-server`, `internal/*` behavior, or tests. Minimal manifest/documentation/example updates (`go.mod`, `go.sum`, `.env.example`, README) are allowed only as needed for relay/Compose usability.

## Consequences

### Positive

- Enables downstream event consumption via Kafka without coupling to or modifying the existing telemetry-server process.
- Keeps failure domains separate and preserves ADR 0004's single-process ingest+serve design for the existing server.
- Provides a reproducible local Kafka environment for development and validation.

### Negative

- Adds an external dependency and operational components (Kafka broker, additional process).
- Relay and telemetry-server may independently consume the same upstream stream, which can duplicate upstream connections/work.
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
- `internal/relay/relay.go`
- `docker-compose.yml`
- `docs/adr/0004-resilient-single-process-ingest-and-serve.md`
