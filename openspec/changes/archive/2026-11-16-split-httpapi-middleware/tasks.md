# Tasks

## 1. Extract middleware package structure

- [x] 1.1 Create `internal/httpapi/middleware/` (package `middleware`) with `middleware.go` containing shared primitives/helpers (`Middleware`, `Chain`, status recorder, request ID context key/helpers, and shared JSON/error helper(s) needed by middleware) moved from `internal/httpapi/middleware.go` with minimal logic changes; verify the package compiles in isolation with `go test ./internal/httpapi/middleware` or `go build ./internal/httpapi/middleware`.
- [x] 1.2 Add `internal/httpapi/middleware/request_id.go`, `logging.go`, `recovery.go`, `body_limit.go`, `timeout.go`, and `ratelimit.go`, moving each middleware concern (including `RateLimiter`/rate-limit helpers) into its corresponding file with behavior preserved; verify the expected files exist and `go build ./internal/httpapi/middleware` succeeds.

## 2. Wire callers and migrate tests

- [x] 2.1 Update `internal/httpapi/server.go` to import/use `internal/httpapi/middleware` (`Chain`, `NewRateLimiter`, and middleware constructors) with the same middleware order and configuration values as before; verify `go build ./internal/httpapi ./cmd/telemetry-server` succeeds.
- [x] 2.2 Move/update `internal/httpapi/middleware_test.go` to test the new `middleware` package (same coverage/expectations for recovery, request ID, and rate limiting; adjust package/imports/exports as needed) and update any other affected `internal/httpapi` tests; verify `go test -race ./internal/httpapi/...` passes.

## 3. Cleanup and validation

- [x] 3.1 Remove the old `internal/httpapi/middleware.go` and resolve any remaining references/import issues; verify `rg -n "internal/httpapi/middleware\\.go|package httpapi" internal/httpapi/middleware` shows the new package files only and `go build ./cmd/...` succeeds.
- [x] 3.2 Run formatting and quality checks for modified code (`gofmt -w .`, `go test -race ./internal/httpapi ./cmd/telemetry-server ./e2e`, `golangci-lint run ./internal/httpapi ./cmd/telemetry-server ./e2e`) and confirm results are clean/passing with no HTTP behavior or OpenAPI contract changes.
