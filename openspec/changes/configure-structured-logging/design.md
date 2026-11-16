# Design

## Context

See `proposal.md` (Why/What) and `specs/structured-logging/spec.md` for motivation and behavioral requirements. The server currently initializes a fixed `slog.TextHandler` in `cmd/telemetry-server/main.go`, loads only `TELEMETRY_WS_URL`/`HTTP_ADDR` in `internal/config`, and has mixed context-aware/non-context-aware log call sites. Request IDs are stored in request context by HTTP middleware but are not consistently attached to log records.

**Assumptions (diagram/rigor):** existing system, plain Mermaid (`flowchart`) for portability, lightweight C4-inspired container view focused on logging/config touchpoints. No deployment or detailed code-level diagram needed.

```mermaid
flowchart LR
    subgraph env[Process Environment / CLI]
      cfgIn[APP_ENV, LOG_LEVEL, LOG_FORMAT, existing flags/env]
    end

    subproc[Telemetry Server Process]
      subgraph cfg[internal/config]
        load[Load + validate + derive defaults]
      end
      subgraph logc[Logging setup cmd/internal logging support]
        init[Build slog handler JSON/Text + level]
        ctxh[Context-aware handler/enrichment adds request_id from ctx]
      end
      subgraph http[internal/httpapi]
        mw[Middleware + Handler]
      end
      subgraph ingest[internal/ingest]
        ing[Ingest + reconnect loop]
      end
      subgraph buf[internal/buffer]
        rb[Ring buffer]
      end
    end

    ext[Upstream Telemetry WebSocket] --> ing
    ing --> rb
    mw --> rb
    cfgIn --> load --> init --> ctxh
    ctxh --> mw
    ctxh --> ing
    mw --> out[stdout logs: JSON in prod / text in dev]
    ing --> out
```

## Goals / Non-Goals

**Goals:**
- Centralize logging runtime config (`APP_ENV`, `LOG_LEVEL`, `LOG_FORMAT`, flags) with fail-fast validation and defaults per spec.
- Initialize one application logger (JSON/text + level) and set it as default for consistent emission.
- Ensure application logs are context-aware and include `request_id` when present in context.
- Keep changes scoped to config, logger setup, call sites, tests, docs, and ADR-0005 (approved).

**Non-Goals:**
- No changes to telemetry buffering, ingest reconnect semantics, or HTTP API request/response contracts.
- No external logging/telemetry backend integration, log shipping, or schema redesign beyond required correlation/config behavior.
- No unrelated refactors of package structure or non-logging behavior.

## Decisions

### 1. Extend `internal/config` as the validation boundary
- **Decision:** Add `AppEnv`, `LogLevel`, and `LogFormat` (or normalized equivalents) to `Config`; parse env/flags in `Load`, normalize case, derive default format from `APP_ENV`, validate allowed values, and return errors on invalid input.
- **Rationale:** Aligns with project rule to validate at extremes, preserves existing flag-over-env behavior, and makes config behavior unit-testable.
- **Alternatives:** Validate in `main` (splits validation, weaker reuse/tests); accept invalid values with fallback (violates agreed fail-fast/spec).

### 2. Build `slog` handler/level in startup composition
- **Decision:** Construct logger in `cmd/telemetry-server/main.go` (or a small helper in an existing/new internal package if needed for reuse/tests) using `slog.NewJSONHandler` or `slog.NewTextHandler` with `slog.HandlerOptions{Level: ...}` based on loaded config, then `slog.SetDefault`.
- **Rationale:** Keeps dependency injection/composition in `cmd` consistent with architecture guidelines while ensuring a single configured logger for injected and default usage.
- **Alternatives:** Per-package logger construction (drift/inconsistency); global mutable config reads inside handlers (harder to test).

### 3. Add a context-aware `slog.Handler` wrapper for request correlation
- **Decision:** Wrap the base handler to extract `request_id` from `context.Context` (reusing existing request ID context key/helpers) and append it as a structured attribute when present on `Handle`. Update application log sites to prefer `DebugContext/InfoContext/WarnContext/ErrorContext` with the relevant `ctx`.
- **Rationale:** Satisfies "all logs need context assigned" and request correlation consistently without changing every call signature beyond passing context; works for request and non-request contexts.
- **Alternatives:** Manually add `"request_id"` args at every call site (error-prone/incomplete); store logger in context only (misses default logger paths). A small wrapper is localized and testable.

### 4. Propagate context through logging call sites
- **Decision:** Audit and update `cmd`, `internal/httpapi` (middleware/handler/server), and `internal/ingest` to use context-aware logging APIs where logs are emitted. Use request context in HTTP paths and run/shutdown contexts in lifecycle/ingest paths.
- **Rationale:** Makes correlation and cancellation/runtime context available to handlers uniformly; meets spec/AGENTS policy.
- **Alternatives:** Leave helper methods contextless and rely only on wrapper (insufficient because non-`*Context` methods don't pass ctx to handler).

### 5. Record ADR-0005
- **Decision:** Draft and, with explicit confirmation at write time, create `docs/adr/0005-configure-structured-logging.md` and update `docs/adr/README.md` documenting env/level/format/context-aware logging policy.
- **Rationale:** Durable architectural/operational decision (approved in discussion); keeps ADRs aligned with implementation policy.
- **Alternatives:** Skip ADR (conflicts with ADR rules for significant decisions); update existing ADR (none cover this concern).

## Risks / Trade-offs

- **[Risk]** Call sites missed during migration remain non-context-aware or omit `request_id`. → **Mitigation:** Search for `logger.Info/Warn/Error/Debug(` and `slog.Info/Warn/Error/Debug(` usages; add/update tests for config parsing and request log correlation where practical.
- **[Risk]** Log format change affects humans/tools expecting text in production today. → **Mitigation:** Default derived from `APP_ENV` with `LOG_FORMAT` override for controlled rollout; document in `.env.example`/README.
- **[Risk]** Invalid env values now fail startup. → **Mitigation:** Intentional fail-fast per spec; clear error messages and documented allowed values.
- **[Trade-off]** Handler wrapper adds slight complexity vs plain handler. → **Mitigation:** Keep wrapper minimal (single responsibility: context attribute extraction).

## Migration Plan

1. Extend config types/parsing/tests and `.env.example`.
2. Add logger construction and context-aware handler wrapper; wire in startup.
3. Update log call sites to pass context / `*Context` methods.
4. Update README and docs as needed; draft ADR-0005 and write only after explicit confirmation.
5. Verify with `gofmt`, `go test -race ./...`, and `golangci-lint run ./...` on modified packages.

**Rollback:** Revert logging config/wrapper/call-site changes to restore previous fixed text logging; no data migration involved.

## Open Questions

- None currently. (Environment variable names, defaults, precedence, fail-fast behavior, and context-correlation scope were resolved during grilling; ADR-0005 creation is approved in principle and requires explicit write-time confirmation.)