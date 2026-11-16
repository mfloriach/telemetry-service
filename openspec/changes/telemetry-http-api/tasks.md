# Tasks

## 1. Project scaffolding and configuration

- [x] 1.1 Create service entrypoint and internal package layout (e.g. `cmd/telemetry-server` and `internal/` packages for config, buffer, ingest, and HTTP) and verify the project builds with `go build ./cmd/...`.
- [x] 1.2 Implement runtime configuration loading for `TELEMETRY_WS_URL` (required) and `HTTP_ADDR` (default `:8080`) with optional CLI flag overrides, and verify unit tests cover defaults, env/flag precedence, and missing required URL behavior.
- [x] 1.3 Confirm and add an approved WebSocket client dependency (per project policy) behind a small internal interface for dialing/reading messages, and verify `go mod tidy` (if needed) and `go build ./...` succeed with the chosen dependency.

## 2. Telemetry buffer and query logic

- [x] 2.1 Implement an in-memory ring buffer that stores the latest 1000 telemetry records as original JSON payloads with an extracted payload `ts` sort key, and verify unit tests cover append order, overwrite-at-capacity behavior, and concurrent read/write safety with `go test -race`.
- [x] 2.2 Implement query logic to return the latest `n` buffered records sorted in descending order by payload `ts` (including deterministic handling when `ts` is missing/uncomparable as defined in implementation), and verify unit tests cover ordering, pass-through record content, and "fewer than n available" cases.
- [x] 2.3 Document the buffer/query package behavior and key assumptions (capacity 1000, in-memory only, `ts` ordering) in code comments or package docs as appropriate, and verify the documented behavior matches implemented tests.

## 3. HTTP API

- [x] 3.1 Implement `GET /telemetry/latest` handler with strict `n` parsing/validation (`1 <= n <= 100`, missing/non-integer/out-of-range -> HTTP 400) and JSON array responses, and verify HTTP handler tests cover valid requests, 400 cases, pass-through JSON objects, descending `ts` order, and partial-result responses.
- [x] 3.2 Wire the HTTP server to use configurable listen address (`HTTP_ADDR`, default `:8080`) and expose the handler, and verify a server-level test confirms the endpoint is reachable and returns expected status/content type for a valid request.
- [x] 3.3 Add or update minimal usage/run documentation (e.g. README section or command docs) describing endpoint, `n` rules, config env vars/flags, and verify the documented invocation/behavior matches the implementation.

## 4. WebSocket ingestion and resilience

- [x] 4.1 Implement the upstream ingestion loop that connects to the configured WebSocket URL, reads JSON telemetry messages, and appends records to the buffer, and verify tests using a fake/test WebSocket server or interface fakes confirm received records are buffered as pass-through JSON.
- [x] 4.2 Implement automatic reconnect behavior on dial/read/connection failures with exponential backoff and jitter (capped, reset on successful reconnect) and structured logs for connect/disconnect/retry events, and verify tests cover retry scheduling/backoff reset behavior and that failures trigger reconnect attempts.
- [x] 4.3 Add focused tests/docs notes for instability handling expectations (periodic drops, best-effort ingest, no persistence across restart), and verify the tests and notes align with the spec scenarios for reconnect and in-memory buffering.

## 5. Service wiring and integration

- [x] 5.1 Wire configuration, buffer, ingestion worker, and HTTP server together in the command entrypoint with context-based startup/shutdown coordination, and verify `go build ./cmd/...` succeeds and the service starts with valid configuration.
- [x] 5.2 Perform end-to-end integration verification against a local/test WebSocket source (including simulated disconnect/reconnect and buffered queries) to confirm `GET /telemetry/latest?n=` returns the latest records sorted by payload `ts` descending, and verify `go test -race ./...` and `golangci-lint run ./...` pass on modified packages.

## Workflow follow-up

- Implement this change with `/opsx-apply` after planning is complete.
- Verify implementation against the change artifacts and archive with `/opsx-archive` once implementation and tracked work are complete.
