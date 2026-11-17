# Design

## Context

The system receives telemetry data over WebSocket connections but lacks a mechanism to forward those records to Kafka. This design adds a relay component that reads telemetry messages from WebSocket and publishes them to Kafka for downstream processing. See proposal.md - Why for motivation.

Key constraints:
- WebSocket dialer interface must be injectable for testing and flexibility
- Kafka publisher interface must be injectable for alternative implementations
- The relay loop must gracefully handle context cancellation and deadline exceeded errors
- Message delivery must be at-least-once, with retry on transient failures

## Goals / Non-Goals

**Goals:**
- Relay telemetry records from WebSocket to Kafka queue
- Provide injectable interfaces for dialer and publisher
- Ensure reliable message delivery with retry semantics
- Maintain existing WebSocket session handling without modification

**Non-Goals:**
- Implement Kafka consumer logic (handled by separate consumer change)
- Define WebSocket message formatting (handled by upstream systems)
- Add authentication/authorization to WebSocket connections (out of scope)

## Decisions

### Decision: Runner loop pattern
- **Chosen**: `runOnce` → `run` loop pattern with `runOnce` handling a single connection cycle
- **Alternative considered**: Single long-running connection with automatic reconnection
- **Rationale**: The `runOnce` pattern allows clean error handling and per-cycle resource management (dial, read, publish, close). The outer `run` loop handles restart logic and context cancellation.

### Decision: Publisher interface
- **Chosen**: `kafka.Publisher` interface with `Publish(ctx, data)` method
- **Alternative considered**: Direct Kafka producer usage within relay
- **Rationale**: The interface abstraction enables unit testing with a mock publisher and supports future substitution of Kafka with another messaging system.

### Decision: Buffer interface in ingest
- **Chosen**: `Buffer` interface with `Add(payload []byte)` method in `internal/ingest`
- **Alternative considered**: Direct buffer implementation without interface
- **Rationale**: The interface allows different buffering strategies (in-memory, persistent) and facilitates testing.

## Risks / Trade-offs

- [Risk] WebSocket connection drops could cause message loss during relay
  - [Mitigation] The ingest buffer stores messages until Kafka publisher is available; retry logic ensures at-least-once delivery
- [Risk] Kafka producer backpressure could block the WebSocket read loop
  - [Mitigation] The buffer decouples WebSocket read speed from Kafka write speed; buffer size should be configured based on expected traffic
- [Risk] Context cancellation timing could leave messages in-flight
  - [Mitigation] The runner checks ctx.Err() before each cycle and ensures connection cleanup in defer