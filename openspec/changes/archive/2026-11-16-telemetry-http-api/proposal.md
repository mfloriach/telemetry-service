# Proposal

## Why

External systems need reliable access to recent device telemetry, but the upstream device WebSocket feed is unstable (connections may fail or drop during periodic instability windows). This change introduces a local HTTP service that continuously ingests telemetry, tolerates disconnects through automatic reconnection and buffering, and serves the latest records on demand.

## What Changes

- Add a long-lived WebSocket ingestion client that connects to a single configurable upstream telemetry URL (e.g. `TELEMETRY_WS_URL`, optional CLI flag override) and consumes JSON telemetry records (~10/sec).
- Add automatic reconnection with exponential backoff and jitter (e.g. ~100ms start, capped at a few seconds), resetting on successful reconnect, with basic structured logs for connect/disconnect/retry events to handle unstable periods.
- Add an in-memory ring buffer retaining the latest **1000** telemetry records (no persistence across restarts) to absorb disconnects and make recent data queryable.
- Add HTTP endpoint `GET /telemetry/latest?n=<int>`:
  - `n` must be an integer in `1 ≤ n ≤ 100`; invalid, missing, or out-of-range values return **HTTP 400**.
  - Returns a JSON array of the latest `n` buffered records sorted in **descending order by payload `ts`**.
  - Returns each record as the original JSON object received (pass-through, all fields preserved).
  - If fewer than `n` records are buffered, returns all available records (still descending by `ts`).
- Add configurable HTTP listen address (e.g. `HTTP_ADDR` env/flag) defaulting to `:8080`.

## Capabilities

### New Capabilities
- `telemetry`: Ingest device telemetry from an unstable WebSocket source with resilient reconnection and buffering, and serve the latest telemetry records over an HTTP query API.

### Modified Capabilities
- None.

## Impact

- **Code:** New Go service components for WebSocket ingestion/reconnect logic, in-memory ring buffer/storage, HTTP server/handler, and configuration/startup wiring (likely under `cmd/` and `internal/` per project layout).
- **APIs:** New external HTTP API `GET /telemetry/latest`.
- **Dependencies:** May require adding a WebSocket client dependency (subject to project approval) and standard library HTTP/JSON/concurrency primitives.
- **Operations/observability:** Runtime configuration for upstream WebSocket URL and HTTP listen address; structured logs for connection lifecycle and retry behavior.
- **Systems:** Acts as a buffering relay between an unstable device telemetry source and downstream consumers. Buffer is in-memory only (lost on restart).
