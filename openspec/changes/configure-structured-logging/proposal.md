# Proposal

## Why

Logging is currently fixed to text output with no level or environment-based configuration, and log calls are not consistently context-aware. This makes production diagnostics harder to ingest/query and weakens request correlation. We need configurable, environment-appropriate structured logging aligned with the updated project logging policy.

## What Changes

- Add runtime logging configuration via environment variables and CLI flags: `APP_ENV` (default `development`), `LOG_LEVEL` (default `info`, `debug|info|warn|error`), and `LOG_FORMAT` (`json|text`, default derived from `APP_ENV`).
- Select JSON logs when `APP_ENV` is `production` (case-insensitive) and text logs otherwise, unless `LOG_FORMAT` explicitly overrides.
- Add matching flags (`-log-level`, `-log-format`) with flag-over-env precedence and fail-fast validation for invalid level/format at startup.
- Initialize `slog` with the configured handler/level and update application log sites to be context-aware (`*Context` logging or equivalent enrichment) so logs include context (including `request_id` when present).
- Update `.env.example`, docs, and tests for logging configuration behavior.
- Record the decision in **ADR-0005** (new; approved by user) after drafting/confirming ADR content.

## Capabilities

### New Capabilities
- `structured-logging`: Configure and emit application logs with environment-aware format, configurable level, and context-correlated structured fields.

### Modified Capabilities
- None.

## Impact

- **Affected code:** `internal/config`, `cmd/telemetry-server/main.go`, `internal/httpapi` (middleware/handler/server), `internal/ingest`, and any other application log call sites.
- **Configuration/docs:** `.env.example`, `README.md`, and related config/logging tests.
- **Architecture records:** new `docs/adr/0005-*.md` and `docs/adr/README.md` index entry (pending explicit ADR-file write confirmation).
- **Dependencies/APIs:** no external dependency changes expected; runtime behavior/log output format becomes configurable by environment. No HTTP API contract changes.
