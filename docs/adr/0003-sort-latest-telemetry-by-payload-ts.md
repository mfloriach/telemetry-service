# ADR-0003: Sort latest telemetry responses by payload ts

**Date**: 2026-11-16
**Status**: accepted
**Deciders**: Project maintainer; coding agent

## Context

`GET /telemetry/latest?n=` must return the latest telemetry records as a JSON array sorted in descending order by reception time/`ts` as clarified during design (payload `ts`), while preserving each original JSON object. Upstream delivery may be unstable and arrival order may not always match payload timestamps; some records may have missing or non-comparable `ts` values.

## Decision

Select up to `n` most recently ingested buffered records and return them sorted in descending order by payload `ts` at read time. Extract a numeric `ts` key on ingest when possible; if `ts` is missing or not comparable as a JSON number, assign a deterministic low key (`math.MinInt64`) so those records sort after valid timestamps.

## Alternatives Considered

### Alternative 1: Sort by service receive time
- **Pros**: Independent of payload clock quality or missing fields.
- **Cons**: Conflicts with the agreed API behavior to sort by payload `ts`.
- **Why not**: The confirmed requirement specifies descending order by payload `ts`.

### Alternative 2: Assume ingestion/arrival order equals payload ts order
- **Pros**: No read-time sorting needed.
- **Cons**: Not guaranteed for unstable or out-of-order upstream data.
- **Why not**: Could violate the API ordering contract.

### Alternative 3: Maintain a globally ts-sorted structure on every insert
- **Pros**: Query reads could avoid sorting work.
- **Cons**: More complex write-path data structures for limited benefit at current scale.
- **Why not**: Sorting at most 1000 buffered records (or requested subset) on read is simple and efficient enough.

## Consequences

### Positive
- API responses consistently follow the agreed payload-`ts` ordering contract.
- Deterministic handling avoids undefined behavior for malformed/missing timestamps.
- Implementation remains simple and easy to test.

### Negative
- Records with invalid/missing `ts` are deprioritized in results, which may surprise consumers expecting those records first/last by other criteria.
- Requires per-ingest `ts` extraction logic and defined fallback behavior.

### Risks
- Ambiguous `ts` formats beyond supported JSON numbers may be treated as uncomparable — mitigated by explicit deterministic fallback and tests.
