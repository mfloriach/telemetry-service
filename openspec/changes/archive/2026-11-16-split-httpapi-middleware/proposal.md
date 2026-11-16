# Proposal

## Why

`internal/httpapi/middleware.go` currently contains all HTTP middleware concerns (request ID, logging, recovery, body limit, rate limiting, timeout) plus shared helpers in one file. Splitting it into a focused package layout improves readability, ownership, and maintainability while preserving existing API behavior.

## What Changes

- Introduce `internal/httpapi/middleware` (package `middleware`) and move middleware code out of `internal/httpapi/middleware.go`.
- Organize as **one file per middleware concern**:
  - `middleware/middleware.go` — shared primitives/helpers (e.g. `Middleware`, `Chain`, status recorder, request ID context/helpers, shared response helpers as needed)
  - `middleware/request_id.go`
  - `middleware/logging.go`
  - `middleware/recovery.go`
  - `middleware/body_limit.go`
  - `middleware/timeout.go`
  - `middleware/ratelimit.go` — rate limiter and rate-limit middleware
- Update `internal/httpapi/server.go`, middleware tests, and imports/usages to the new package layout.
- Remove/replace the old monolithic `internal/httpapi/middleware.go`.
- Behavior-preserving refactor only: no endpoint, response, validation, timeout, rate-limit, logging, or OpenAPI behavior changes intended.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- None. This change has no spec-level behavior changes (`skip_specs: true` in `.openspec.yaml`).

## Impact

- **Affected code:** `internal/httpapi/middleware.go`, new `internal/httpapi/middleware/*`, `internal/httpapi/server.go`, `internal/httpapi/middleware_test.go` (and any related imports in `internal/httpapi` / tests).
- **APIs/specs:** no external HTTP API or OpenAPI contract changes.
- **ADRs:** reviewed `docs/adr/0001`–`0004`; no existing ADR requires change and no new ADR is required.
- **Risks:** import cycles or accidental behavior drift during extraction; mitigated by keeping logic identical and running `gofmt`, `go test -race`, and `golangci-lint` on modified packages.
