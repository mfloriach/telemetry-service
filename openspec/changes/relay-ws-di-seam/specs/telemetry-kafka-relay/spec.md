# Spec Delta

## MODIFIED Requirements

### Requirement: Relay ingests telemetry from a configured upstream WebSocket

The relay SHALL read telemetry records from an injected WebSocket connection provided by the composition root and publish each record to Kafka.

#### Scenario: Successful record intake from injected WebSocket connection

- **WHEN** the relay receives a connected WebSocket from the caller
- **THEN** it reads incoming JSON telemetry records and publishes each to the configured Kafka topic

### Requirement: Relay restores upstream connectivity after instability

The composition root SHALL re-dial the upstream WebSocket and inject a new connection when the relay reports a connection loss; the relay SHALL resume publishing on the new connection without restarting the process.

#### Scenario: Upstream disconnect during operation

- **WHEN** the injected WebSocket connection is lost
- **THEN** the relay stops publishing and returns the read error; the composition root re-dials and injects a new connection