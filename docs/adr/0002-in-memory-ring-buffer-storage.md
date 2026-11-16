# ADR-0002: Store recent telemetry in an in-memory ring buffer

**Date**: 2026-11-16
**Status**: accepted
**Deciders**: Project maintainer; coding agent

## Context

External consumers need recent telemetry via `GET /telemetry/latest?n=1..100`, but the upstream feed can drop during instability windows. The service must buffer received records for querying, return original JSON objects (pass-through), and remain simple to operate. Durability across process restarts was explicitly out of scope.

## Decision

Use a concurrency-safe, in-memory fixed-capacity ring buffer (default/latest 1000 records) that stores each record as the original JSON payload plus an extracted `ts` sort key, overwriting the oldest records when full.

## Alternatives Considered

### Alternative 1: Persistent storage (database/file/queue)
- **Pros**: Data survives restarts; supports historical queries and stronger durability guarantees.
- **Cons**: Adds operational and implementation complexity not required by the API or agreed scope.
- **Why not**: Persistence was explicitly not required; restart loss was accepted.

### Alternative 2: Unbounded in-memory buffering
- **Pros**: Retains all received records while running.
- **Cons**: Memory usage can grow without bound with large/high-rate payloads.
- **Why not**: A bounded buffer provides predictable memory and is sufficient for serving `n <= 100`.

### Alternative 3: Store decoded maps/structs only
- **Pros**: Convenient typed access to fields.
- **Cons**: Risk of losing/coercing original JSON fields, conflicting with pass-through response requirements.
- **Why not**: Preserving the original payload is required for response fidelity.

## Consequences

### Positive
- Predictable memory footprint and simple implementation.
- Pass-through responses preserve all original JSON fields.
- Fast local reads/writes suitable for ~10 records/sec ingest and small query windows.

### Negative
- Buffered data is lost on restart.
- Capacity is bounded by record count (1000), not total bytes.

### Risks
- Large individual payloads could increase memory use despite count cap — mitigated by documented scope assumptions and bounded record count.
