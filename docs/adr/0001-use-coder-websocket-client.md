# ADR-0001: Use coder/websocket for upstream telemetry ingestion

**Date**: 2026-11-16
**Status**: superseded by [ADR-0009](0009-relay-hosts-websocket-ingest-server.md)
**Deciders**: Project maintainer; coding agent

## Context

The telemetry service must ingest JSON records from a single upstream device WebSocket endpoint that is unstable and periodically disconnects. Implementing WebSocket dialing, framing, read loops, and close behavior directly on raw sockets would add avoidable complexity and risk in reconnect-heavy code, while the project prefers standard-library-first solutions and requires explicit approval before adding external dependencies.

## Decision

Use `github.com/coder/websocket` as the WebSocket client dependency, with the client in `pkg/ws` and the consumer-side `Read`/`Close` interface in `internal/relay` so ingestion and reconnect logic remain testable with fakes.

## Alternatives Considered

### Alternative 1: github.com/gorilla/websocket
- **Pros**: Mature, widely used, well-known API and ecosystem.
- **Cons**: Different API style; not the approved choice for this change.
- **Why not**: The maintainer explicitly approved `coder/websocket` for this implementation.

### Alternative 2: Standard-library-only custom WebSocket implementation
- **Pros**: Avoids adding a third-party dependency.
- **Cons**: Significantly higher implementation and correctness burden (handshake, framing, control frames, close semantics) for an unstable connection path.
- **Why not**: Higher risk and maintenance cost outweighed dependency avoidance for this scope.

## Consequences

### Positive
- Reliable, maintained WebSocket client behavior for connect/read/close flows.
- Interface boundary enables deterministic unit tests without requiring a real network WebSocket client in every test.

### Negative
- Adds an external dependency to `go.mod`/`go.sum`, requiring dependency governance.
- Future upgrades or behavior changes in the library must be tracked.

### Risks
- Dependency/API drift over time — mitigated by isolating usage behind `pkg/ws` and covering behavior with tests.

## References

- `pkg/ws/ws.go` — WebSocket client and connection options.
- `internal/relay/relay.go` — consumer-side `Read`/`Close` interface (removed by ADR-0009).
