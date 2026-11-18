# Design

## Context

See proposal.md - Why for motivation. The current state: `cmd/relay/main.go` cannot build because it passes `ws.CoderDialer{}` (a `ws.Dialer`) where `relay.NewRunner` expects a `relay.Conn`. A dialer creates connections; it is not a connection. Additionally, `internal/relay/relay.go` accepts a `wsURL` parameter but never dials it — the `Dial` call is commented out, so `runner.conn` is never initialized and `wsURL` is unused.

The `pkg/ws` package owns the WebSocket connection type (`coderConn`) but that type is unexported, so `main.go` cannot name it in a struct field. And `coderConn.Read` is a generic method (`Read[M any]`), which cannot satisfy `relay.Conn`'s non-generic `Read(ctx) (Message, error)`. A small contract must cross the package boundary without importing either direction.

Key constraints:
- `internal/relay` must NOT import `pkg/ws` (no coupling FROM relay TO ws)
- `pkg/ws` must NOT import `internal/relay` (no coupling FROM ws TO relay)
- `cmd/relay/main.go` is the composition root and may import both
- `relay.Run` no longer takes `wsURL`; the URL is dialed in `main`

## Goals / Non-Goals

**Goals:**
- Fix the build failure in `cmd/relay`
- Remove all coupling between `pkg/ws` and `internal/relay` (zero import edges either way)
- Preserve end-to-end behavior: relay reads WebSocket telemetry, publishes to Kafka
- Keep `relay.Conn` and `kafka.Publisher` interfaces injectable for testability

**Non-Goals:**
- Changing `internal/relay/relay.go` behavior beyond removing `wsURL` and the commented-out dial
- Adding reconnection/backoff logic inside `relay` (moved to `main`)
- Modifying `cmd/telemetry-server`, `internal/utils/kafka`, or `internal/utils/buffer`

## Decisions

### Decision: `ws` exports a `Conn` interface; `Dial` returns it
- **Chosen**: `pkg/ws` defines `type Conn interface { Read(ctx) (websocket.MessageType, []byte, error); Close(code, reason) error }` and `Dialer.Dial` returns `Conn`
- **Alternative considered**: `internal/relay` defines its own `Dialer` and `Conn`, and `pkg/ws` implements them by importing `relay`
- **Rationale**: The latter couples `ws` → `relay`, violating the no-coupling constraint. `ws` owns the connection type, so it should declare the contract. `main.go` adapts `ws.Conn` → `relay.Conn`.

### Decision: `main.go` adapts `ws.Conn` → `relay.Conn` via a small adapter struct
- **Chosen**: `main.go` defines an unexported adapter type holding a `ws.Conn`, with a `Read` method that converts `websocket.MessageType` → `relay.Message` and delegates `Close`
- **Alternative considered**: `ws` defines `Read` returning `relay.Message` directly
- **Rationale**: That would couple `ws` → `relay`. Keeping the adapter in `main` keeps both packages decoupled.

### Decision: `relay.Run(ctx)` drops `wsURL`; `main` dials and injects
- **Chosen**: `NewRunner(conn Conn, publisher Publisher)` unchanged; `Run(ctx) error` takes no URL; `main` calls `dialer.Dial(ctx, url)`, adapts the result, passes to `NewRunner`
- **Alternative considered**: `relay` keeps a `Dialer` field and dials internally (current broken state)
- **Rationale**: Current state is broken (Dial commented out, `conn` nil). Moving dial to `main` fixes the build and keeps `relay` focused on the read/publish loop.

### Decision: Connection re-establishment handled by `main`, not `relay`
- **Chosen**: When `relay.Run` returns a read error, `main` re-dials the WebSocket, adapts the new connection, and calls `Run` again in a loop
- **Alternative considered**: `relay.Run` contains its own dial-and-reconnect loop
- **Rationale**: `relay` should not know how to dial; `main` owns connection lifecycle per the DI seam decision

## Risks / Trade-offs

- [Risk] `main.go` adapter struct is unexported and only usable within `main`
  - [Mitigation] Acceptable — the adapter is composition-root glue, not a reusable component
- [Risk] `ws.Conn.Read` returns `websocket.MessageType` while `relay.Message.Type` is `relay.MessageType`; adapter must convert
  - [Mitigation] Simple map in adapter's `Read` method; covered by unit tests
- [Risk] `main` must implement reconnection loop; relay no longer retries internally
  - [Mitigation] Spec requirement explicitly places reconnection on the composition root; tests will cover the loop

## Open Questions

- Should `main`'s reconnection loop use bounded backoff, or immediate retry? The relay spec defers this to the composition root; a decision is needed before implementation.
- Does `cmd/relay/main.go` need to handle `Dial` errors separately from `Run` read errors, or treat both as "reconnect"?
