# ADR-0004: Run resilient ingest and HTTP serving in a single process

**Date**: 2026-11-16
**Status**: accepted
**Deciders**: Project maintainer; coding agent

## Context

The system acts as a buffering relay between an unstable upstream WebSocket feed (~10 records/sec with periodic disconnects) and downstream HTTP consumers requesting recent records. The solution must automatically reconnect, log connection lifecycle/retry events, and serve `GET /telemetry/latest` from local buffered state with minimal operational complexity.

## Decision

Implement a single Go service process that runs a resilient WebSocket ingestion worker and an HTTP server concurrently, sharing an in-memory buffer. The ingestion loop reconnects on dial/read/connection failures using exponential backoff with jitter (capped), and resets backoff after a healthy successful session; startup/shutdown is coordinated with `context.Context` and OS signals.

## Alternatives Considered

### Alternative 1: Separate ingest and query services with shared persistent storage
- **Pros**: Independent scaling/lifecycle and durable shared state.
- **Cons**: Requires external storage and distributed coordination, increasing complexity beyond scope.
- **Why not**: Buffering is explicitly in-memory and the workload is small; a split architecture is unnecessary.

### Alternative 2: Fixed-delay reconnect retries
- **Pros**: Simple retry logic.
- **Cons**: Less adaptive during sustained instability and can create avoidable retry pressure.
- **Why not**: Exponential backoff with jitter better matches unstable upstream behavior.

### Alternative 3: No automatic reconnect / fail-fast process
- **Pros**: Simpler control flow.
- **Cons**: Service becomes unavailable during expected periodic upstream instability.
- **Why not**: Continuous availability of recent buffered/query service despite upstream drops is a core requirement.

## Consequences

### Positive
- Simple deployment and operation (one binary/process, local state).
- Resilient ingestion during unstable periods with observable structured logs.
- Clear shutdown behavior via context cancellation and HTTP server shutdown.

### Negative
- Ingest and query availability are coupled in one process/runtime.
- Horizontal scaling and durable state are not addressed by this design.

### Risks
- A process crash/restart loses buffered data and interrupts serving — mitigated by accepting in-memory scope and documenting restart behavior.
- Mis-tuned backoff could delay recovery or retry too aggressively — mitigated by capped exponential backoff with jitter and tests covering reconnect/reset behavior.
