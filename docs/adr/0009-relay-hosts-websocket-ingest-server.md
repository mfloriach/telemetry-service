# ADR-0009: Relay hosts the device WebSocket ingest server

**Date**: 2026-11-22
**Status**: accepted
**Deciders**: Project maintainer; coding agent

## Context

The relay (`cmd/relay`) was a WebSocket client: it dialed a single upstream URL from `TELEMETRY_WS_URL`, read records, and published them to Kafka. The telemetry server consumed Kafka and served HTTP. Nothing in the system listened for devices, so the local stack had no way to receive device telemetry and the HTTP API stayed empty.

Two responsibilities were also conflated in the old model: "connect to somebody else's feed" and "accept telemetry from our devices". Devices need a stable endpoint, concurrent connections, and a clear failure signal when their record is not accepted.

## Decision

Make the relay the single WebSocket ingest point and reverse the transport role of `pkg/ws` for this flow:

- **Server:** add `pkg/ws/server.go` with a `Server` that accepts handshakes on the fixed path `/telemetry`, listens on `WS_ADDR` (default `:9000`), enforces a 1 MiB default read limit, and closes connections with an explicit status and reason. The existing client in `pkg/ws/ws.go` stays available for other uses but is no longer used by the relay.
- **Concurrency:** each device connection is handled independently in its own goroutine, and all connections publish through one shared, concurrency-safe Kafka writer.
- **Failure policy:** the server retries through the Kafka writer's own attempts; when a publish ultimately fails, only the affected connection is closed with a server-error status (1011). The relay does not buffer records while Kafka is unavailable; devices reconnect and resend.
- **Boundary:** `pkg/ws` owns transport lifecycle and close semantics; `internal/relay` owns the decision to accept a record and publish it. The relay no longer dials an upstream URL.
- **Configuration:** relay requires `KAFKA_BROKERS`, uses `KAFKA_TOPIC` (default `telemetry.raw`), and uses `WS_ADDR` (default `:9000`). Relay no longer reads `TELEMETRY_WS_URL`.
- **Single ingest:** `cmd/telemetry-server` does not host a WebSocket ingest route; it remains a Kafka consumer with an HTTP read API.

This supersedes ADR-0001 and ADR-0007, which describe `pkg/ws` as a dialing client with a single-caller consumer interface, and amends ADR-0006's relay runtime, configuration, and reconnect statements.

## Alternatives Considered

### Alternative 1: Keep the relay as a client to an external upstream feed

- **Pros**: No new server code; unchanged relay behavior.
- **Cons**: Devices still have no endpoint to connect to, so the system cannot ingest device telemetry.
- **Why not**: It does not solve the missing ingest path.

### Alternative 2: Host the WebSocket ingest inside `cmd/telemetry-server`

- **Pros**: One fewer process boundary for local development.
- **Cons**: Creates two paths into Kafka, couples the HTTP read API to device connection management, and contradicts the existing separation where telemetry-server consumes Kafka.
- **Why not**: Relay already owns Kafka publishing; duplicating ingest there splits responsibility and failure modes.

### Alternative 3: Buffer records in the relay while Kafka is unavailable

- **Pros**: Fewer visible failures for devices during short outages.
- **Cons**: Requires durable or bounded in-memory buffering, retry ordering, and backpressure policy; the Kafka writer already retries transient failures.
- **Why not**: It adds durability semantics the system does not need yet and hides failures from devices that could simply retry.

### Alternative 4: Make the WebSocket path configurable

- **Pros**: More deployment flexibility.
- **Cons**: Another required setting that all callers must agree on, with no current need.
- **Why not**: A fixed `/telemetry` path keeps the device contract simple; the listen address remains configurable.

## Consequences

### Positive

- Devices have a concrete endpoint: `ws://<host>:<WS_ADDR>/telemetry`.
- Concurrent device connections are handled independently, and one failing device does not disturb the others.
- Transport behavior (accept, limits, close codes, shutdown) is testable in isolation from Kafka.
- The relay is the single ingest authority, so there is one path into Kafka.

### Negative

- **BREAKING:** `TELEMETRY_WS_URL` is no longer relay configuration; deployments and local `.env` files must switch to `WS_ADDR`.
- A Kafka outage now closes affected device connections instead of hiding behind a server-side buffer.
- `pkg/ws` contains both a client and a server, and the client is currently unused by production code.

### Risks

- Device fleets that do not retry will lose records during a Kafka outage — mitigated by the explicit 1011 close and documented retry expectation.
- The ingest endpoint has no TLS or authentication — acceptable for local development, but a prerequisite before exposing it outside localhost.
- A wedged connection could delay shutdown — mitigated by closing tracked connections with a going-away status and bounding the HTTP shutdown.

## References

- `pkg/ws/server.go` — WebSocket server, handler contract, close taxonomy.
- `pkg/ws/ws.go` — retained WebSocket client.
- `internal/relay/handler.go` — publish handler shared by all connections.
- `internal/relay/config.go` — `WS_ADDR` configuration and validation.
- `cmd/relay/main.go` — server startup, signal handling, publisher lifecycle.
- `openspec/changes/relay-websocket-server/` — proposal, specs, design, and tasks for this decision.
