# Telemetry Server

HTTP API service that ingests satellite telemetry from an unstable WebSocket source, buffers recent records in memory, and serves the latest telemetry over HTTP.

## Configuration

Copy `.env.example` to `.env` and adjust values as needed.

The server requires an upstream WebSocket URL and supports configurable runtime settings for HTTP and logging.

- `TELEMETRY_WS_URL` (required): upstream WebSocket telemetry source URL
- `HTTP_ADDR` (optional): HTTP listen address, default `:8080`
- `APP_ENV` (optional): application environment, default `development`. When set to `production` (case-insensitive) and `LOG_FORMAT` is not explicitly set, logs default to JSON; otherwise logs default to text
- `LOG_LEVEL` (optional): log level, default `info`. Allowed values: `debug`, `info`, `warn`, `error` (case-insensitive)
- `LOG_FORMAT` (optional): log output format, default derived from `APP_ENV` (`production` => `json`, otherwise `text`). Allowed values: `json`, `text` (case-insensitive)

CLI flags override environment variables:

- `-telemetry-ws-url=<url>`
- `-http-addr=<addr>`
- `-app-env=<env>`
- `-log-level=<level>`
- `-log-format=<format>`

Invalid `LOG_LEVEL` or `LOG_FORMAT` values cause startup to fail with a configuration error. Application logs are emitted with runtime/request context; request-scoped logs include `request_id` when available.

## Run

```bash
# using environment variables
TELEMETRY_WS_URL=ws://localhost:9000/telemetry go run ./cmd/telemetry-server

# with CLI flags (flags override env vars)
TELEMETRY_WS_URL=ws://localhost:9000/telemetry go run ./cmd/telemetry-server -http-addr=:8080 -log-level=debug

# explicit flags (including logging config)
go run ./cmd/telemetry-server \
  -telemetry-ws-url=ws://localhost:9000/telemetry \
  -http-addr=:8080 \
  -app-env=production \
  -log-level=info \
  -log-format=json
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

## Development

### Hot reload (Air)

Install [Air](https://github.com/air-verse/air) and run:

```bash
air
```

Air uses `.air.toml` to rebuild and restart `./cmd/telemetry-server` on changes.

### Pre-commit hooks (Lefthook)

This project uses [Lefthook](https://github.com/evilmartians/lefthook) (Go-based git hooks manager) to run formatting checks, linting, and tests before commits.

```bash
# install lefthook (example)
go install github.com/evilmartians/lefthook@latest

# install git hooks
lefthook install
```

Pre-commit runs `gofmt` checks, `golangci-lint run ./...`, and `go test -race ./...` (see `lefthook.yml`). Ensure `golangci-lint` is installed locally.

### Docker

Build and run with the multi-stage Dockerfile (uses BuildKit cache mounts):

```bash
docker build -t telemetry-server .
docker run --rm -p 8080:8080 \
  -e TELEMETRY_WS_URL=ws://host.docker.internal:9000/telemetry \
  telemetry-server
```

### Local Kafka (Docker Compose) and relay

A local single-broker Kafka (KRaft) service is provided for development:

```bash
docker compose up -d
# Broker for host clients: localhost:9092
```

Run the relay (`cmd/relay`) to consume telemetry from the upstream WebSocket and publish each JSON record to Kafka (existing telemetry-server behavior is unchanged):

```bash
# required for relay
export TELEMETRY_WS_URL=ws://localhost:9000/telemetry
export KAFKA_BROKERS=localhost:9092
# optional (default: telemetry.raw)
export KAFKA_TOPIC=telemetry.raw

go run ./cmd/relay
```

Optional logging env vars (`APP_ENV`, `LOG_LEVEL`, `LOG_FORMAT`) follow the same conventions as the telemetry server. Startup fails with a clear error if required relay configuration is missing or invalid.

### CI

GitHub Actions (`.github/workflows/ci.yml`) runs formatting checks, `golangci-lint`, `go test -race ./...`, a build, and a Docker image build with GitHub Actions cache on pushes and pull requests.

## Notes

- Buffering is in-memory only (latest 1000 records); data is not persisted across restarts.
- Upstream ingestion automatically reconnects with exponential backoff + jitter during unstable periods.
- Telemetry payloads are treated as pass-through JSON; `ts` is used for ordering.
