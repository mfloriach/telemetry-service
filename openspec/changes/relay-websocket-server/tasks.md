# Tasks

## 1. WebSocket server in `pkg/ws`

- [x] 1.1 Add `pkg/ws/server.go` with `Server`, `ServerOption`, `MessageHandler`, `NewServer`, `WithServerReadLimit`, `WithServerLogger`, and `Addr`, validating a non-empty address, a non-nil handler, and a positive read limit; verify `go build ./pkg/ws/...` succeeds and `pkg/ws/ws.go` is unchanged (`git diff --stat pkg/ws/ws.go` shows nothing)
- [x] 1.2 Implement `Server.Run(ctx)` with `net.Listen` first, an `http.ServeMux` exposing only `GET /telemetry`, `websocket.Accept`, tracked live connections, per-message handler invocation, and the close taxonomy from design §5 (normal close, mapped read error, 1011 handler failure, 1001 shutdown); verify the server tests added in 1.3 pass
- [x] 1.3 Add table-driven `pkg/ws/server_test.go` covering address/read-limit/handler validation, rejection of non-`/telemetry` paths, text and binary round-trips, default and custom read limits, oversize-message rejection without handler invocation, handler-error close 1011, graceful shutdown close 1001, and a nil logger; verify `go test -race ./pkg/ws`
- [x] 1.4 Update the package comment so `pkg/ws` describes both client and server behavior; verify `gofmt -l pkg/ws` is empty and `golangci-lint run ./pkg/ws/...` passes

## 2. Relay publish handler and configuration

- [x] 2.1 Add `internal/relay/handler.go` with `Publisher` and `Handler.Handle(ctx, messageType, payload)` that publishes payload bytes unchanged and wraps publish failures; verify table-driven `internal/relay/handler_test.go` covers success, frame-type passthrough, payload fidelity, and error propagation with `go test -race ./internal/relay`
- [x] 2.2 Remove `internal/relay/relay.go` (`Runner`, `Conn`) now that multi-connection handling lives in `pkg/ws`; verify `go build ./...` succeeds and no remaining references exist (`grep -rn "NewRunner\|relay.Conn" --include="*.go" .`)
- [x] 2.3 Replace `TelemetryWSURL` with `WSAddr` in `internal/relay/config.go`: default `:9000`, env `WS_ADDR`, flag `-ws-addr`, `net.SplitHostPort` shape validation, and removal of the old required-field error; verify `go test -race ./internal/relay`
- [x] 2.4 Rewrite `internal/relay/config_test.go` as table-driven cases for default address, env value, flag precedence, malformed address, missing `KAFKA_BROKERS`, and absence of any `TELEMETRY_WS_URL` requirement; verify `go test -race ./internal/relay`
- [x] 2.5 Update the relay section of `.env.example` to document `WS_ADDR` (optional, default `:9000`) instead of `TELEMETRY_WS_URL`, leaving local `.env` untouched; verify `grep -n "TELEMETRY_WS_URL" .env.example` returns nothing and `grep -n "WS_ADDR" .env.example` shows the documented default

## 3. Relay wiring

- [x] 3.1 Update `cmd/relay/main.go` to build the publish handler and `ws.NewServer(cfg.WSAddr, handler.Handle, ws.WithServerLogger(logger))`, derive a signal-aware context, log the bound address, run the server, and close the Kafka publisher only after the server stops; verify `go build ./cmd/relay` and `golangci-lint run ./cmd/relay/... ./internal/relay/...` pass
- [x] 3.2 Verify startup failures surface clearly by running the built relay without `KAFKA_BROKERS` and with `WS_ADDR=not-an-address`; both commands exit non-zero with messages naming the offending value

## 4. Telemetry server cleanup

- [x] 4.1 Remove the leftover `pkg/wsingest` directory and confirm no `/ws` route or publisher wiring remains in `cmd/telemetry-server` or `internal/telemetry`; verify `grep -rn "\"/ws\"\|wsingest" --include="*.go" .` returns nothing and `go build ./cmd/telemetry-server` succeeds
- [x] 4.2 Run `make smoke` and confirm the telemetry server still serves `GET /telemetry/latest` and shuts down cleanly with only HTTP and Kafka configuration

## 5. Documentation and ADRs

- [x] 5.1 Add `docs/adr/0009-relay-hosts-websocket-ingest-server.md` recording the server decision, the concurrent-connection and publish-failure policies, and the config/endpoint change; verify the file follows the existing ADR format (date, status, deciders, context, decision, alternatives, consequences)
- [x] 5.2 Mark ADR-0001 and ADR-0007 as superseded by ADR-0009, amend ADR-0006's runtime/config and reconnect statements, and update the status/rows in `docs/adr/README.md`; verify each ADR's status line and the README table agree (`grep -n "0009\|superseded" docs/adr/*.md`)
- [x] 5.3 Update `docs/architecture.md` to the device → relay → Kafka → server flow: replace the upstream-source actor, use current package paths (`pkg/ws`, `pkg/kafka`, `internal/telemetry/*`), update the component table and `TelemetryWSURL` discussion, and keep the Mermaid container diagram consistent with design §8; verify the document contains no `TELEMETRY_WS_URL` or `internal/utils/` references (`grep -n "TELEMETRY_WS_URL\|internal/utils" docs/architecture.md` returns nothing)
- [x] 5.4 Update `README.md`: start commands use `WS_ADDR` for relay, examples connect devices to `ws://localhost:9000/telemetry`, remove stale telemetry-server `TELEMETRY_WS_URL` usage, and correct the reconnect note to say devices retry; verify every documented relay command runs as written (`go run ./cmd/relay` with the listed env vars starts and fails fast only on missing `KAFKA_BROKERS`)

## 6. Tooling and end-to-end verification

- [x] 6.1 Point the `Makefile` `device` target at `ws://localhost:9000/telemetry` and update the relay example in `docker-compose.yml` comments to `WS_ADDR`; verify `make -n device` prints the new URL and `docker compose config --quiet` succeeds
- [ ] 6.2 Run `gofmt -l .`, `go build ./cmd/...`, `go test -race ./...`, and `golangci-lint run ./...`; all must be clean
- [ ] 6.3 With `docker compose up -d`, `go run ./cmd/relay`, and `make device` running, confirm `curl "http://localhost:8080/telemetry/latest?n=3"` returns the records sent by the device, and that stopping the relay logs the listen address and exits cleanly on SIGINT

## Workflow follow-up

- Archive the change with `/opsx-archive` after review and after `openspec validate "relay-websocket-server" --strict` passes.
