# Tasks

## 1. Export Conn interface from pkg/ws

- [x] 1.1 Add exported `Conn` interface to `pkg/ws/ws.go` (`Read(ctx) (websocket.MessageType, []byte, error)`, `Close(code, reason)`)
- [x] 1.2 Change `CoderDialer.Dial` return type from `*coderConn` to `Conn`
- [x] 1.3 Verify `coderConn` satisfies `ws.Conn` and package compiles: `go build ./pkg/ws/...`

## 2. Remove wsURL from relay.Run and commented-out dial

- [x] 2.1 Remove `wsURL string` parameter from `relay.Run(ctx)` in `internal/relay/relay.go`
- [x] 2.2 Remove commented-out `Dial` block from `runOnce`/`Run` in `internal/relay/relay.go`
- [x] 2.3 Verify `internal/relay` compiles: `go build ./internal/relay/...`

## 3. Wire ws.Conn → relay.Conn in cmd/relay/main.go

- [x] 3.1 Add adapter struct in `cmd/relay/main.go` holding `ws.Conn`, implementing `relay.Conn` (Read converts `websocket.MessageType` → `relay.Message`, Close delegates)
- [x] 3.2 Dial WS URL in `main` via `ws.CoderDialer.Dial(ctx, cfg.TelemetryWSURL)`
- [x] 3.3 Inject adapted `Conn` into `relay.NewRunner(adapter, publisher)` instead of `ws.CoderDialer{}`
- [x] 3.4 Call `runner.Run(ctx)` without `wsURL`
- [x] 3.5 Add reconnection loop in `main`: on `Run` read error, re-dial, re-adapt, re-run
- [x] 3.6 Verify `cmd/relay` builds: `go build ./cmd/relay/...`

## 4. Verification

- [ ] 4.1 Write unit test for `ws.Conn` adapter Read/Close conversion in `cmd/relay` or `pkg/ws`
- [ ] 4.2 Run full project build and tests: `go build ./... && go test -race ./...`
- [ ] 4.3 Confirm `cmd/relay` binary builds and starts without dial errors

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.