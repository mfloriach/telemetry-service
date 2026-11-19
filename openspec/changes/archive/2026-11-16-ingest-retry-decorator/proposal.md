# Proposal

## Why

`internal/ingest/ingest.go` currently combines single-session WebSocket ingestion and resilient reconnect/backoff+jitter policy in one `Ingester` type, making retry behavior harder to understand and evolve in isolation. Refactoring to a decorator separates concerns while preserving the existing resilient ingestion behavior required by ADR-0004 and the `telemetry` spec.

## What Changes

- Refactor ingest implementation to a decorator-style structure:
  - Introduce a base/session ingester responsible for one connection/read session (`RunOnce` semantics: dial, read, buffer records, return on disconnect/cancel).
  - Introduce a retry decorator (e.g. resilient/retrying runner) that wraps the base ingester and owns reconnect loop policy: exponential backoff with jitter (capped), delay/sleep hook, and backoff reset after a healthy successful session.
- Move retry policy configuration/hooks (`baseDelay`, `maxDelay`, `jitter`, `sleep`, `successMin`) to the retry decorator and update `internal/ingest` tests/callers to the new internal API.
- Update `cmd/telemetry-server/main.go` and ingest tests (including reconnect tests) to compose base ingester + retry decorator with equivalent runtime behavior.
- Keep behavior-preserving refactor only: no changes intended to reconnect semantics, buffering/query behavior, logging intent, or HTTP behavior.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- None. This change has no spec-level behavior changes (`skip_specs: true` in `.openspec.yaml`).

## Impact

- **Affected code:** `internal/ingest/ingest.go`, `internal/ingest/ingest_test.go`, `internal/ingest/reconnect_test.go`, `cmd/telemetry-server/main.go` (and any related internal call sites).
- **APIs:** internal `ingest` package API/types/constructors will change; external HTTP API and telemetry behavior remain unchanged.
- **ADRs/specs:** reviewed `docs/adr/0004` and `openspec/specs/telemetry/spec.md`; compatible as behavior-preserving, no ADR/spec changes required.
- **Risks:** accidental retry-policy drift or test brittleness during refactor; mitigated by preserving existing behavior, updating deterministic tests, and running `gofmt`, `go test -race`, and `golangci-lint` on modified packages.
