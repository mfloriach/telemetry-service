# ADR-0005: Configure structured logging by environment and level

**Date**: 2026-11-16
**Status**: accepted
**Deciders**: Project maintainer; coding agent

## Context

The telemetry server previously emitted fixed text logs with no runtime level or environment-based format configuration, and application log sites were not consistently context-aware. As operational needs and project guidance evolved, logs needed to be machine-friendly in production, readable in development, configurable by level, and consistently correlated with runtime/request context (such as request IDs) to support diagnosis during unstable upstream conditions.

## Decision

We configure application logging at startup from validated runtime settings: `APP_ENV` (default `development`), `LOG_LEVEL` (default `info`, `debug|info|warn|error`), and `LOG_FORMAT` (`json|text`, default derived from `APP_ENV` where `production` selects JSON and all other environments select text). CLI flags (`-app-env`, `-log-level`, `-log-format`) take precedence over environment variables, and invalid logging configuration fails fast before serving traffic. The service uses a configured `slog` handler and requires context-aware logging so request/runtime context is propagated and `request_id` is included in log records when present.

## Alternatives Considered

### Alternative 1: Keep fixed text logging with minimal ad hoc fields
- **Pros**: No configuration surface or startup validation changes; simple implementation.
- **Cons**: Poor fit for production log ingestion/querying and inconsistent request correlation.
- **Why not**: It does not meet the need for environment-appropriate structured logs and context-assigned logging.

### Alternative 2: Hardcode JSON in production without configurable level/format overrides
- **Pros**: Ensures structured production logs with less configuration complexity.
- **Cons**: Limits operational control (e.g. debug-level troubleshooting, temporary format overrides) and makes local/test behavior less flexible.
- **Why not**: The service needs explicit, validated runtime control over log level and format while keeping safe defaults.

### Alternative 3: Adopt a third-party logging framework
- **Pros**: Potentially richer features and ecosystem integrations.
- **Cons**: Adds dependency and migration cost for requirements already covered by the standard library.
- **Why not**: Go's `slog` provides sufficient structured logging, handler/level configuration, and extensibility for this scope.

## Consequences

### Positive
- Production logs are structured (JSON) by default while development remains human-readable.
- Operators can control verbosity and format via validated environment/CLI configuration.
- Context-aware logging improves traceability across HTTP requests and runtime/ingestion flows.

### Negative
- Startup is stricter and will fail on invalid logging configuration values.
- Logging call sites and tests require ongoing discipline to pass context and preserve correlation fields.

### Risks
- Some log call sites may be missed during migration, reducing correlation consistency — mitigated by auditing `slog`/logger call sites, using a context-aware handler wrapper, and adding targeted tests.
- Environment/default mismatches could surprise operators — mitigated by explicit defaults, fail-fast validation, and documentation in `.env.example` and `README.md`.
