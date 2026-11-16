# Tasks

## 1. Local Kafka environment

- [x] 1.1 Add `docker-compose.yml` with a single-broker Kafka service in KRaft mode for local development, including broker listener/client connectivity settings and minimal usage comments; verify with `docker compose config` succeeding and the documented service/config keys present.
- [x] 1.2 Add or update minimal relay/Compose usage notes in `README.md` (and `.env.example` only if needed for relay variables) without changing existing telemetry-server runtime behavior; verify the documented Kafka/relay variables and startup steps match the compose and relay config names.

## 2. Relay command and configuration

- [x] 2.1 Add `cmd/relay/main.go` with a standalone entrypoint that loads required runtime configuration (`TELEMETRY_WS_URL`, `KAFKA_BROKERS`, `KAFKA_TOPIC` with default `telemetry.raw` as specified) and optional logging environment settings, failing startup with clear errors for missing/invalid required config; verify `go build ./cmd/relay` succeeds and invalid/missing-config runs exit non-zero with actionable error output.
- [x] 2.2 Add any relay-specific supporting packages/code (without modifying existing `internal/*` behavior) for upstream WebSocket connection, bounded reconnect/backoff handling, and Kafka publishing of each received JSON telemetry record as the original message value; verify with `gofmt -w .`/`go build ./cmd/relay` and targeted tests or a runnable local check showing reconnect attempts and publish path behavior.

## 3. Dependencies and architecture record

- [x] 3.1 Add required Kafka client dependency updates in `go.mod`/`go.sum` only as needed for the relay, keeping existing application dependencies/behavior intact; verify `go mod tidy` (if applicable) and `go build ./cmd/...` succeed.
- [x] 3.2 Create the approved ADR for Kafka relay/Compose/separate process using the next available ADR number at write time and update `docs/adr/README.md` consistently; verify ADR numbering, title, status/date, links, and decision content match this change and do not conflict with existing ADRs.

## 4. Validation

- [x] 4.1 Run relevant checks for modified/new files (`gofmt -w .` and `gofmt -l .`, `go build ./cmd/...`, and `go test -race ./...` or affected package tests as applicable) and manually validate relay flow with local Kafka (`docker compose up -d`, run `cmd/relay` against a telemetry source, confirm published messages preserve original telemetry records on the configured topic); verify results and confirm existing telemetry-server behavior/tests remain unchanged.

## Workflow follow-up

- Verify implementation against `proposal.md`, `specs/telemetry-kafka-relay/spec.md`, and this design before archive.
- Archive the change with `/opsx-archive` after implementation and tracked work are complete, recording verification results and final commit message recommendation if needed.
