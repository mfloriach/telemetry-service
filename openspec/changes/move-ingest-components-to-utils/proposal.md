# Proposal

## Why

This change refactors the WebSocket ingestion components to decouple the connection interface from the implementation. By moving the `Conn`, `Dialer`, and `Message` types to `internal/utils/` and having `SessionIngester` depend on the `Dialer` interface rather than a concrete websocket implementation, the system can swap WebSocket libraries or migrate to a different WebSocket technology without modifying the ingester core. Currently, `internal/ingest/ingest.go` and `internal/ingest/ws.go` contain concrete WebSocket types that make the ingester tightly coupled to a specific WebSocket library. This change enables technology flexibility and improves testability.

## What Changes

- Move `MessageType`, `Message`, `Conn` interface, and `Dialer` interface from `internal/ingest/ws.go` to `internal/utils/`
- Refactor `SessionIngester` in `internal/ingest/ingest.go` to depend on `Dialer` interface instead of concrete connection type
- Update `internal/ingest/retry.go` to use the shared `utils.WithRetry` mechanism
- The ingester's `RunOnce` method will accept any `Dialer` implementing the interface

## Capabilities

### New Capabilities

- `ingest/websocket-interface`: A cohesive system behavior that owns the WebSocket connection contract. This capability can own related requirements as the system evolves. Name uses kebab-case following the project's existing spec organization. Each creates `specs/ingest/websocket-interface/spec.md`.

### Modified Capabilities
- (None - this is a pure interface extraction refactoring)

## Impact

- **Code**: `internal/ingest/ingest.go`, `internal/ingest/ws.go` → `internal/utils/`
- **Interfaces**: `Conn`, `Dialer`, `Message`, `MessageType` types moved to shared utility
- **Dependencies**: Other packages depending on `internal/ingest` types will need to update imports
- **Testing**: Interfaces make the ingester more testable with mock connections

## Commands

```bash
# After proposal, create specs:
openspec instructions specs --change "move-ingest-components-to-utils" --json

# Then design:
openspec instructions design --change "move-ingest-components-to-utils" --json
```
