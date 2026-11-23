# ADR-0007: Configure the WebSocket connection with functional options and a required endpoint

**Date**: 2026-11-19
**Status**: superseded by [ADR-0009](0009-relay-hosts-websocket-ingest-server.md)
**Deciders**: Project maintainer; coding agent

## Context

The upstream telemetry WebSocket client was an empty `CoderDialer` struct that called `websocket.Dial(ctx, url, nil)`. Every connection therefore ran on the library defaults: a 32 KiB read limit that rejects larger telemetry frames, no handshake deadline of its own, no compression negotiation, and no way to attach authentication headers. With no fields on the dialer, the relay had no lever to change any of it.

The client also has exactly one caller in this repository, so an incompatible redesign costs nothing beyond updating that call site.

## Decision

Build connections through `ws.NewConn(ctx, url string, opts ...Option)`. The endpoint is a required positional argument, so the compiler enforces it, and every other setting is a functional option: handshake timeout, read limit, compression, subprotocols, headers, and logger. The constructor validates its configuration and fails before dialing. Reconnect policy stays with the caller rather than moving into the transport.

## Alternatives Considered

### Alternative 1: A `Config` struct parameter
- **Pros**: Familiar shape; fields are self-documenting.
- **Cons**: The zero value is ambiguous — a `0` read limit and "unset" look identical, forcing pointer fields or sentinel values throughout validation.
- **Why not**: Ambiguity moves the defaulting problem into every caller instead of keeping it in one constructor.

### Alternative 2: Pass `websocket.DialOptions` straight through
- **Pros**: No new option types; direct access to every library knob.
- **Cons**: Leaks the library's types into the package contract, pins callers to the library, and still leaves the read limit, handshake deadline, and logger to solve separately.
- **Why not**: The package would expose its implementation rather than its own configuration surface.

### Alternative 3: Automatic reconnection inside the connection type
- **Pros**: One place handles disconnects.
- **Cons**: Hides lifecycle and retry policy in the transport; the caller can no longer decide when a dead connection ends the run.
- **Why not**: Reconnection is caller policy, and folding it in would change relay behavior outside this decision's scope.

## Consequences

### Positive
- One construction path carries every configurable setting, with defaults that are safe for telemetry traffic.
- Invalid configuration surfaces at startup with an error naming the offending setting, before any network activity.
- The consumer-side `Read`/`Close` interface in `internal/relay` is unchanged, so the relay runner and its fakes are unaffected.

### Negative
- The package's API is now breaking: `CoderDialer` is removed and the only caller must switch to `NewConn`.
- Compression, headers, and subprotocols are opt-in, so callers that need them must say so explicitly.
- A default 1 MiB read limit admits frames the previous 32 KiB limit rejected.

### Risks
- Added options can express contradictory settings — mitigated by constructor validation, such as rejecting a compression threshold while compression is disabled.
- A misconfigured endpoint fails at startup rather than degrading gracefully — mitigated by logging the endpoint and error through the injected logger.
- Callers may expect the transport to reconnect — mitigated by documenting that reconnect policy belongs to the caller.

## References

- `pkg/ws/ws.go` — connection constructor, options, and defaults.
- `openspec/changes/ws-conn-configurable-constructor/design.md` — decision record for this change.
