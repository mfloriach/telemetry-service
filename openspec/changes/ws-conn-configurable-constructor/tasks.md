# Tasks

## 1. Connection API and construction

- [x] 1.1 Replace `CoderDialer`/`coderConn` in `pkg/ws/ws.go` with an exported `Conn` type, an `Option` type, and the `NewConn(ctx, url string, opts ...Option)` signature; verify `go build ./pkg/ws` succeeds and `rg 'CoderDialer|coderConn' pkg/ws` returns nothing (the `cmd/relay` reference is replaced in 3.1).
- [x] 1.2 Define the option constructors (`WithHandshakeTimeout`, `WithReadLimit`, `WithCompression`, `WithSubprotocols`, `WithHeader`, `WithLogger`) and the defaults (1 MiB read limit, 10s handshake deadline, compression off, no extra subprotocols or headers, nil logger) so options apply over defaults; verify with table-driven tests in `pkg/ws/ws_test.go` asserting each option changes only its own field and an optionless construction yields the documented defaults.
- [x] 1.3 Implement edge validation in `NewConn` for an empty URL, a non-positive handshake timeout, a non-positive read limit, and a compression threshold supplied while compression is disabled; verify table-driven tests assert each case returns an error naming the offending setting, and use an `httptest` server with a request counter to prove no dial is attempted.
- [x] 1.4 Wire the options into the dial: handshake deadline on the connect context, read limit on the connection, compression (enabled flag plus threshold), request headers, offered subprotocols, and the injected logger on failure; verify an `httptest` WebSocket server that asserts the received headers and negotiated subprotocol, plus a slow-handshake server whose connect fails within the configured deadline rather than hanging.

## 2. Read and close behavior

- [x] 2.1 Keep `Read(ctx)` returning the frame type as a `string` alongside the payload, and surface oversize messages as an error with no partial payload; verify tests against a test server sending a message at and above the configured limit assert the full payload and the error respectively.
- [x] 2.2 Keep `Close(code int, reason string) error` sending the code and reason and releasing the connection even when the close handshake fails; verify tests observe the close code and reason on the server side and that the underlying connection is released in both the success and failure paths.
- [x] 2.3 Emit dial and close failures through the configured logger and stay silent when none is set; verify a `slog.Logger` writing to a buffer captures the endpoint and error for an unreachable URL, while construction without a logger produces no output.

## 3. Relay call site

- [x] 3.1 Replace the `ws.CoderDialer{}.Dial(...)` call in `cmd/relay/main.go` with `ws.NewConn(ctx, cfg.TelemetryWSURL, ws.WithLogger(logger))`, leaving `relay.Load` and the runner call untouched; verify `go build ./cmd/...` succeeds and `git diff cmd/relay/main.go` shows only that call-site change.
- [x] 3.2 Confirm the consumer contract is unchanged: `internal/relay` still declares its own two-method `Conn` interface and `relay.NewRunner` receives `*ws.Conn` without a conversion; verify `go test -race ./internal/relay` passes with no edits to that package.

## 4. Decision records

- [x] 4.1 Amend `docs/adr/0001-use-coder-websocket-client.md` so the boundary reads `pkg/ws` for the client with the consumer-side `Read`/`Close` interface in `internal/relay`, update the references and the risk line to match, and keep the status `accepted`; verify `rg 'internal/ingest' docs/adr/0001-use-coder-websocket-client.md` returns nothing and every path the ADR names exists in the repository.
- [x] 4.2 Add `docs/adr/0007-configure-websocket-connection-with-functional-options.md` using the existing template, recording the functional-options decision, the rejected alternatives, and that reconnect policy stays with the caller; add its row to `docs/adr/README.md`; verify the index link resolves to the new file and its listed title and status match the file header.

## 5. Integration verification

- [x] 5.1 Run `gofmt -l pkg/ws/ws.go pkg/ws/ws_test.go cmd/relay/main.go`, `go build ./cmd/...`, `go test -race ./pkg/ws ./internal/relay`, and `golangci-lint run ./pkg/ws/... ./cmd/relay/...`; verify all report clean, and `git status --short` lists only `pkg/ws/ws.go`, `pkg/ws/ws_test.go`, `cmd/relay/main.go`, the `docs/adr/` files, and this change's `openspec/` directory, with no changes under `internal/relay`, `go.mod`, or `go.sum`. Repository-wide `gofmt` and tests have pre-existing failures at HEAD (`internal/telemetry/config/config.go` formatting and `internal/telemetry/httpapi` tests) that this change does not touch.

## Workflow follow-up

- Archive the change with `/opsx-archive` once implementation is complete and reviewed.
- Correct the `internal/utils/ws` path in the `architecture.md` component table in a separate docs-only change.
