# Tasks

## 1. Refactor ingest into session + retry decorator

- [x] 1.1 Refactor `internal/ingest/ingest.go` to introduce a base/session ingester that implements single-session ingestion (`RunOnce`: dial, read, buffer, close, return ctx/disconnect errors) and a retry decorator that wraps it to implement resilient `Run` with exponential backoff+jitter (capped), context-aware sleep, and backoff reset after healthy successful sessions; move retry policy/config hooks (`baseDelay`, `maxDelay`, `jitter`, `sleep`, `successMin`) to the decorator with defaults preserving current behavior, and verify `go build ./internal/ingest`.
- [x] 1.2 Update `internal/ingest/ingest_test.go` to exercise the base/session ingester `RunOnce` behavior (buffering, dial/read errors, cancellation) against the new API while preserving assertions/intent, and verify `go test -race ./internal/ingest -run RunOnce`.

## 2. Update callers and reconnect tests

- [x] 2.1 Update `internal/ingest/reconnect_test.go` to target the retry decorator (compose base + decorator, configure deterministic policy hooks on decorator) and preserve reconnect/backoff/reset behavior assertions, and verify `go test -race ./internal/ingest -run RunReconnect`.
- [x] 2.2 Update `cmd/telemetry-server/main.go` (and any other internal call sites) to construct and run the composed base ingester + retry decorator with equivalent behavior, and verify `go build ./cmd/telemetry-server ./internal/ingest`.

## 3. Validation

- [x] 3.1 Run formatting and quality checks for modified code (`gofmt -w .`, `go test -race ./internal/ingest ./cmd/telemetry-server ./e2e`, `golangci-lint run ./internal/ingest ./cmd/telemetry-server ./e2e`) and confirm results are clean/passing with no behavior changes to reconnect semantics or HTTP/OpenAPI behavior.
