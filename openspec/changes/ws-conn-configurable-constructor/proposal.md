# Proposal

## Why

`pkg/ws` dials the upstream telemetry source with `websocket.Dial(ctx, url, nil)`, so every connection inherits coder/websocket's defaults: a 32 KiB read limit that fails on larger telemetry frames, no handshake deadline of its own, no compression negotiation, and no way to attach authentication headers. The relay — the package's only caller — has no lever to change any of it, because `CoderDialer` is an empty struct.

## What Changes

- Add `NewConn(ctx context.Context, url string, opts ...Option) (*Conn, error)` to `pkg/ws`. The URL is the single mandatory parameter, passed as a concrete `string`; every other setting is a functional option.
- Options: `WithHandshakeTimeout`, `WithReadLimit`, `WithCompression`, `WithSubprotocols`, `WithHeader`, `WithLogger`.
- Defaults: 1 MiB read limit (up from coder's 32 KiB), 10s handshake timeout, compression disabled, no extra subprotocols or headers — so today's behavior holds apart from accepting larger frames.
- Validate at the edge: empty URL, non-positive timeout, non-positive read limit, and a compression threshold set while compression is disabled all fail in `NewConn`.
- Return an exported `*Conn` implementing `Read` and `Close`; the consumer-side interface stays in `internal/relay`.
- **BREAKING**: remove `CoderDialer`. `cmd/relay/main.go` constructs the connection through `NewConn` with `relay.Config`'s URL, and logs the dial result.
- Explicitly out of scope: reconnection/backoff, any `internal/relay` change, and new env vars or flags.

## Capabilities

### New Capabilities
- `telemetry/ws-connection`: the upstream WebSocket client contract — a mandatory URL plus configurable handshake timeout, read limit, compression, subprotocols, headers, and logging, exercised through `Read` and `Close`.

### Modified Capabilities
<!-- none: the relay's publish path and runtime configuration are unchanged -->

## Impact

Code: `pkg/ws/ws.go`, `cmd/relay/main.go`. No dependency changes, no new required configuration, no behavior change for the relay's publish path.

Decisions recorded: ADR-0001 is amended because the `Dialer`/`Conn` boundary it describes moved from `internal/ingest` to `pkg/ws` with the consumer interface in `internal/relay`; a new ADR-0007 records functional options as the configuration pattern and leaves reconnect policy to the caller.
