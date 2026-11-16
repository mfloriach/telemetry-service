# Telemetry Server

HTTP API service that ingests satellite telemetry from an unstable WebSocket source, buffers recent records in memory, and serves the latest telemetry over HTTP.

## Configuration

Copy `.env.example` to `.env` and adjust values as needed.

The server requires an upstream WebSocket URL and supports a configurable HTTP listen address.

- `TELEMETRY_WS_URL` (required): upstream WebSocket telemetry source URL
- `HTTP_ADDR` (optional): HTTP listen address, default `:8080`

CLI flags override environment variables:

- `-telemetry-ws-url=<url>`
- `-http-addr=<addr>`

## Run

```bash
# using environment variables
TELEMETRY_WS_URL=ws://localhost:9000/telemetry go run ./cmd/telemetry-server

# with CLI flags (flags override env vars)
TELEMETRY_WS_URL=ws://localhost:9000/telemetry go run ./cmd/telemetry-server -http-addr=:8080

# explicit flag for upstream URL
go run ./cmd/telemetry-server -telemetry-ws-url=ws://localhost:9000/telemetry -http-addr=:8080
```

Build:

```bash
go build ./cmd/...
```

## HTTP API

OpenAPI specification: [`docs/openapi.yaml`](docs/openapi.yaml)

### `GET /telemetry/latest`

Query parameters:

- `n` (required): integer `1 <= n <= 100`

Behavior:

- Returns `200 OK` with a JSON array of the latest `n` buffered telemetry records.
- Records are sorted in descending order by payload `ts`.
- Each record is returned as the original JSON object received (pass-through).
- If fewer than `n` records are buffered, all available records are returned.
- Invalid, missing, or out-of-range `n` returns `400 Bad Request` with a JSON error body: `{"error":"..."}`.
- Excessive request rates return `429 Too Many Requests` with a JSON error body.
- Request timeouts return `503 Service Unavailable`.
- Unexpected failures return `500 Internal Server Error` with a JSON error body.
- Responses include an `X-Request-Id` correlation header.

Example:

```bash
curl "http://localhost:8080/telemetry/latest?n=5"
```

## Notes

- Buffering is in-memory only (latest 1000 records); data is not persisted across restarts.
- Upstream ingestion automatically reconnects with exponential backoff + jitter during unstable periods.
- Telemetry payloads are treated as pass-through JSON; `ts` is used for ordering.
