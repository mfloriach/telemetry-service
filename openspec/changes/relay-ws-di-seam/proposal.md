# Proposal

## Why

`cmd/relay/main.go` fails to build because it passes `ws.CoderDialer{}` (a `ws.Dialer`) where `relay.NewRunner` expects a `relay.Conn`. A dialer is not a connection — it creates connections. Worse, `internal/relay/relay.go` accepts a `wsURL` but never dials it (the `Dial` call is commented out), so the URL is accepted but unused, and `runner.conn` is never initialized. The relay must dial the upstream WebSocket in `main.go` and inject an already-connected `Conn` into the runner, with `pkg/ws` exporting an interface contract so neither `ws` nor `relay` imports the other.

## What Changes

- **New file**: `pkg/ws/ws.go` — add exported `Conn` interface (`Read(ctx) (websocket.MessageType, []byte, error)`, `Close(code, reason)`); change `Dial` to return `Conn` instead of unexported `*coderConn`
- **Modified file**: `internal/relay/relay.go` — remove `wsURL` parameter from `Run(ctx)`; relay no longer dials, accepts pre-connected `Conn` via `NewRunner`
- **Modified file**: `cmd/relay/main.go` — dial WS URL via `ws.CoderDialer`, adapt `ws.Conn` → `relay.Conn` via a small adapter struct in `main`, inject into `relay.NewRunner`, call `runner.Run(ctx)` without `wsURL`

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- `telemetry-kafka-relay`: Relay connection sourcing changes from internal WebSocket dialing to injected connection; `wsURL` parameter moves from `relay.Run` to `main.go`. End-to-end behavior (ingest WebSocket, publish to Kafka) is unchanged; only the DI seam and connection ownership move.

## Impact

- **Modified code**: `internal/relay/relay.go` (remove `wsURL`, remove commented-out dial), `cmd/relay/main.go` (dial + adapt + inject), `pkg/ws/ws.go` (export `Conn` interface, `Dial` returns `Conn`)
- **Breaking API change**: `relay.Run` signature loses `wsURL`; `ws.CoderDialer.Dial` return type changes from `*coderConn` to `ws.Conn`
- **No impact on**: `internal/telemetry-server`, `internal/utils/kafka`, `internal/utils/buffer`, existing `telemetry-kafka-relay` end-to-end behavior