# Tasks

## 1. Implement relay runner

- [x] 1.1 Implement `internal/relay/relay.go` with `runner` type, `NewRunner`, and `Run` methods
- [x] 1.2 Implement `runOnce` method for single connection cycle with dial, read, publish, close
- [x] 1.3 Ensure context cancellation is checked before each cycle
- [x] 1.4 Verify connection cleanup in defer with proper close frame

## 2. Implement session ingester

- [x] 2.1 Implement `internal/ingest/ingest.go` with `SessionIngester` type, `NewSessionIngester`, and `Run` methods
- [x] 2.2 Implement `buffer` interface with `Add(payload []byte)` method
- [x] 2.3 Implement `Run` method for WebSocket connection and message loop
- [x] 2.4 Handle context cancellation and deadline exceeded errors
- [x] 2.5 Verify message buffering on read errors

## 3. Define interfaces

- [x] 3.1 Implement `internal/utils/ws/dialer.go` with `Dialer` interface
- [x] 3.2 Implement `internal/utils/kafka/publisher.go` with `Publisher` interface
- [x] 3.3 Verify interfaces are injectable for testing

## 4. Integration verification

- [x] 4.1 Write integration test for WebSocket-to-Kafka message flow
- [x] 4.2 Verify at-least-once delivery with retry logic
- [x] 4.3 Test buffer behavior under Kafka unavailability