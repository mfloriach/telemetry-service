# Tasks

## 1. Logging Configuration

- [x] 1.1 Extend `internal/config` (`Config`, `Load`, flags/env parsing) to support `APP_ENV` (default `development`), `LOG_LEVEL` (default `info`, `debug|info|warn|error`), and `LOG_FORMAT` (`json|text`, default derived from `APP_ENV`), including `-log-level`/`-log-format` flags with flag-over-env precedence and fail-fast validation for invalid values; update/add `internal/config` tests to verify defaults, derivation, precedence, normalization, and invalid-input errors via `go test -race ./internal/config`.
- [x] 1.2 Update `.env.example` and `README.md` to document `APP_ENV`, `LOG_LEVEL`, and `LOG_FORMAT` (allowed values/defaults/precedence and production JSON vs dev text behavior) and verify examples/flag names match `internal/config` and `cmd/telemetry-server` usage.

## 2. Logger Initialization and Context Correlation

- [x] 2.1 Implement configured `slog` logger initialization (JSON/text handler + level) and a context-aware handler/enrichment wrapper that adds `request_id` from context when present; wire logger setup in `cmd/telemetry-server/main.go` (and set default logger) so startup uses the validated config, and verify with targeted tests (new or existing package tests) that format/level selection and request ID attribute injection behave as specified.
- [x] 2.2 Draft `docs/adr/0005-configure-structured-logging.md` content and the `docs/adr/README.md` index entry for the approved logging decision; after obtaining explicit write-time confirmation, create/update those ADR files and verify ADR numbering/title/status/links are consistent with `docs/adr/README.md`.

## 3. Context-Aware Log Call Sites

- [x] 3.1 Update `internal/httpapi` logging (middleware/handler/server and related helpers) to use context-aware logging APIs with request context and preserve/emit request correlation fields; update/add `internal/httpapi` tests to verify request-scoped logs include `request_id` and logging behavior remains correct via `go test -race ./internal/httpapi ./internal/httpapi/...`.
- [x] 3.2 Update `internal/ingest` and `cmd/telemetry-server` lifecycle/startup/shutdown log call sites to use context-aware logging with runtime contexts; update/add relevant tests (e.g. `internal/ingest`) to verify behavior and absence of regressions via `go test -race ./internal/ingest ./cmd/...`.

## 4. Integration Verification

- [x] 4.1 Run `gofmt -w .` on modified files and verify `gofmt -l .` reports no files, then run `go test -race ./...` and `golangci-lint run ./...` to confirm the change passes project checks.

## Workflow follow-up

- Implement this change with `/opsx-apply` once planning is approved.
- Verify implementation against artifacts with `/opsx-verify` as needed.
- Archive the change with `/opsx-archive` after implementation and tracked work are complete.
- Create/update ADR-0005 files only with explicit write-time confirmation if not completed during implementation.
