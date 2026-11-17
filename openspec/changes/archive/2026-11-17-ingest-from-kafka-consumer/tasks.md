# Tasks

## 1. Implement Kafka consumer interface

- [x] 1.1 Implement `internal/utils/kafka/consumer.go` with `Consumer` interface (`Consume(ctx) ([]byte, error)`, `Close() error`)
- [x] 1.2 Write unit test for `Consumer` interface contract and verify it compiles
- [x] 1.3 Verify `Consumer` interface is injectable for testing (mock consumer compiles and satisfies interface)

## 2. Modify SessionIngester for Kafka consumer

- [x] 2.1 Modify `internal/ingest/ingest.go` to replace `dialer ws.Dialer` with `consumer kafka.Consumer` struct field
- [x] 2.2 Modify `NewSessionIngester` to accept `kafka.Consumer` instead of `ws.Dialer` and verify constructor compiles
- [x] 2.3 Modify `Run` method signature to `Run(ctx context.Context) error` (remove `url` parameter)
- [x] 2.4 Replace WebSocket `conn.Read(ctx)` loop with Kafka consumer `Consume(ctx)` loop
- [x] 2.5 Verify context cancellation returns `ctx.Err()` from the ingest loop
- [x] 2.6 Verify non-context consumer read errors are wrapped and returned
- [x] 2.7 Verify buffer `Add(payload)` is called with consumed message payloads
- [x] 2.8 Remove `ws` import from `internal/ingest/ingest.go` and verify package compiles

## 3. Update callers and integration verification

- [x] 3.1 Find all callers of `NewSessionIngester` and update them to pass a `kafka.Consumer`
- [x] 3.2 Verify all callers compile after the interface change
- [x] 3.3 Write integration test for Kafka consumer → buffer message flow
- [x] 3.4 Run `go test ./internal/ingest/...` and verify all ingest tests pass