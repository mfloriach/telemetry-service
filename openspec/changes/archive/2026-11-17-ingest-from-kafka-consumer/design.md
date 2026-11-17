# Design

## Context

The `SessionIngester` in `internal/ingest/ingest.go` currently couples ingest logic to WebSocket transport via `ws.Dialer` and `ws.Conn`. It dials a URL, reads messages with `conn.Read(ctx)`, and buffers payloads. See proposal.md - Why for motivation. This design decouples ingest from WebSocket transport by replacing the WebSocket dialer with a Kafka consumer interface, modeled after the existing `kafka.Publisher` pattern in `internal/utils/kafka/publisher.go`.

Key constraints:
- The `kafka.Consumer` interface must follow the same injectable pattern as `kafka.Publisher` for testability
- Context cancellation must stop the ingest loop and return `ctx.Err()`
- Non-context consumer read errors must be wrapped and returned
- Buffer behavior (`Add(payload []byte)`) remains unchanged

## Goals / Non-Goals

**Goals:**
- Decouple `SessionIngester` from WebSocket transport concerns
- Align ingest component with the existing Kafka-based relay architecture
- Maintain injectable interfaces for testability
- Preserve buffer semantics (`Buffer.Add`) unchanged

**Non-Goals:**
- Implementing the concrete Kafka consumer (out of scope; `kafka.Consumer` is an interface)
- Changing `internal/relay/relay.go` (unchanged)
- Changing `internal/utils/ws` (unchanged)
- Implementing consumer group rebalancing logic (handled by the Kafka client library)

## Decisions

### Decision: Consumer interface modeled after Publisher pattern
- **Chosen**: `kafka.Consumer` interface with `Consume(ctx) ([]byte, error)` and `Close() error` methods, mirroring `kafka.Publisher`'s `Publish`/`Close` structure
- **Alternative considered**: Using `kafka.Reader` directly in `SessionIngester`
- **Rationale**: The interface abstraction enables unit testing with mock consumers and follows the same pattern already established by `kafka.Publisher` in the codebase

### Decision: Remove URL parameter from Run
- **Chosen**: `Run(ctx context.Context) error` — no URL parameter
- **Alternative considered**: Keeping URL for consumer bootstrap servers
- **Rationale**: The Kafka consumer is configured externally (brokers, topic, group ID) before being passed to `SessionIngester`; the URL is no longer relevant to ingest logic

### Decision: Buffer remains unchanged
- **Chosen**: `Buffer` interface with `Add(payload []byte)` stays as-is
- **Alternative considered**: Adding a Kafka-specific buffer wrapper
- **Rationale**: The buffer is transport-agnostic; changing it would introduce unnecessary coupling

## Risks / Trade-offs

- [Risk] Callers of `NewSessionIngester` must update to pass a `kafka.Consumer` instead of `ws.Dialer`
  - [Mitigation] Document the breaking change; update all call sites
- [Risk] Kafka consumer group rebalancing may cause temporary pauses in ingestion
  - [Mitigation] Handled by the Kafka client library; spec acknowledges rebalancing behavior
- [Risk] Consumer offset management (commit strategy) is not specified
  - [Mitigation] Deferred to consumer implementation; `SessionIngester` only reads payloads