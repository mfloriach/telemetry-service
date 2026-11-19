# Proposal

## Why

The project ingests upstream device telemetry and serves recent records over HTTP, but has no way to forward ingested telemetry into a durable/event-streaming system for downstream consumers. Adding a Kafka relay and local Compose setup enables event publishing without changing existing telemetry-server behavior.

## What Changes

- Add `docker-compose.yml` for local development with a single-broker Kafka service (KRaft mode) and documented client connectivity.
- Add a new binary at `cmd/relay/main.go` that independently connects to the configured upstream WebSocket telemetry source, reads telemetry JSON records, and publishes each record to Kafka.
- Add relay runtime configuration (env vars / CLI flags consistent with repo conventions): `TELEMETRY_WS_URL`, `KAFKA_BROKERS`, `KAFKA_TOPIC` (default suggestion `telemetry.raw`), and optional `LOG_LEVEL`/`LOG_FORMAT`/`APP_ENV`; include reconnect/backoff handling for upstream instability.
- Keep existing application code, tests, and runtime behavior unchanged. Allow minimal manifest/doc/example updates required for usability (e.g. `go.mod`/`go.sum`, `.env.example`, README notes) only as needed for the new relay/Compose.
- Record an approved architecture decision (new ADR, next available number at write time) for introducing Kafka, local Compose support, and the separate relay process.

## Capabilities

### New Capabilities
- `telemetry-kafka-relay`: Independently relay upstream telemetry JSON records to a configured Kafka topic via a dedicated relay command and local Kafka Compose environment, without modifying existing telemetry-server ingest/serve behavior.

### Modified Capabilities
- None.

## Impact

- **New code:** `cmd/relay/` (and supporting relay-specific packages only if added without modifying existing packages).
- **Infra/dev tooling:** `docker-compose.yml` (Kafka/KRaft, local dev).
- **Dependencies/config/docs:** possible updates to `go.mod`/`go.sum`, `.env.example`, and README for Kafka client dependency and relay/Compose usage.
- **Unchanged:** existing `cmd/telemetry-server`, `internal/*` behavior, existing tests, and current HTTP/WebSocket ingest semantics.
- **ADRs/specs:** new ADR required (approved) for Kafka relay/deployment decision; no changes to existing `telemetry` or `structured-logging` requirement specs expected.
