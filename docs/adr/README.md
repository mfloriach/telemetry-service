# Architecture Decision Records

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [0001](0001-use-coder-websocket-client.md) | Use coder/websocket for upstream telemetry ingestion | superseded by [0009](0009-relay-hosts-websocket-ingest-server.md) | 2026-11-16 |
| [0002](0002-in-memory-ring-buffer-storage.md) | Store recent telemetry in an in-memory ring buffer | accepted | 2026-11-16 |
| [0003](0003-sort-latest-telemetry-by-payload-ts.md) | Sort latest telemetry responses by payload ts | accepted | 2026-11-16 |
| [0004](0004-resilient-single-process-ingest-and-serve.md) | Run resilient ingest and HTTP serving in a single process | accepted | 2026-11-16 |
| [0005](0005-configure-structured-logging.md) | Configure structured logging by environment and level | accepted | 2026-11-16 |
| [0006](0006-kafka-telemetry-relay.md) | Kafka telemetry relay as a separate process | accepted (amended by [0009](0009-relay-hosts-websocket-ingest-server.md)) | 2026-11-16 |
| [0007](0007-configure-websocket-connection-with-functional-options.md) | Configure the WebSocket connection with functional options and a required endpoint | superseded by [0009](0009-relay-hosts-websocket-ingest-server.md) | 2026-11-19 |
| [0008](0008-configure-kafka-clients-with-functional-options.md) | Configure the Kafka clients with functional options and explicit required values | accepted | 2026-11-20 |
| [0009](0009-relay-hosts-websocket-ingest-server.md) | Relay hosts the device WebSocket ingest server | accepted | 2026-11-22 |
| [0010](0010-use-docusaurus-for-docs.md) | Use Docusaurus for the documentation site | proposed | 2026-11-23 |
