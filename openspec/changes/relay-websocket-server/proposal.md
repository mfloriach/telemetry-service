# Proposal

## Why

Devices have no WebSocket endpoint to connect to: `cmd/relay` currently dials an external upstream URL, while the telemetry server only serves HTTP. As a result, the local stack (`docker compose` + `air` + `make device`) cannot deliver device records into Kafka, so the HTTP API stays empty.

## What Changes

- Add a WebSocket server implementation in a new file under `pkg/ws`; keep the existing client implementation intact.
- Start the server from `cmd/relay/main.go` at the fixed path `/telemetry`, listening on `WS_ADDR` (default `:9000`).
- Accept concurrent device connections and publish each received text or binary record unchanged to `KAFKA_TOPIC`.
- Close a device connection when publishing fails, so the device detects the failure and can retry.
- **BREAKING**: relay no longer dials `TELEMETRY_WS_URL`; relay configuration uses `WS_ADDR` instead. The client code remains available but is no longer used by the relay.
- Remove the `telemetry-server` `/ws` handler and its Kafka publisher wiring added in the previous uncommitted work; relay becomes the single WebSocket ingest point.
- Update `.env.example`, the `Makefile` `device` target, `docker-compose.yml` comments, and README to use `ws://localhost:9000/telemetry`.
- Record the reversal in ADR-0009; mark ADR-0001 and ADR-0007 superseded and amend ADR-0006.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `telemetry/ws-connection`: add the server-side accept/read/close contract while preserving the existing client contract.
- `telemetry/relay`: relay accepts multiple device connections and publishes each record; a failed publish closes that connection.
- `telemetry-kafka-relay`: relay configuration and intake change from an upstream URL to a hosted WebSocket endpoint.

## Impact

- **Code**: new server file in `pkg/ws`; `cmd/relay/main.go`; `internal/relay` (configuration and connection handling); remove `pkg/wsingest`, its `telemetry-server` wiring, and its HTTP route.
- **Config/API**: **BREAKING** `TELEMETRY_WS_URL` no longer configures relay; `WS_ADDR` (default `:9000`) and fixed path `/telemetry` are used. No new dependencies.
- **Docs/tooling**: `.env.example`, `Makefile`, `docker-compose.yml`, README, ADR-0009, and affected ADR statuses.
- **Systems**: device → relay (`ws://localhost:9000/telemetry`) → Kafka → telemetry-server → HTTP.
