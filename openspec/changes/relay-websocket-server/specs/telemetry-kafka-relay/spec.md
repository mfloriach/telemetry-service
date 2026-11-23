# Spec Delta

## MODIFIED Requirements

### Requirement: Relay validates required runtime configuration
The relay SHALL use `WS_ADDR` (default `:9000`) as its WebSocket listen address and SHALL require `KAFKA_BROKERS` and `KAFKA_TOPIC` (default `telemetry.raw`). It SHALL fail startup with a clear error when a required value is missing or invalid. `TELEMETRY_WS_URL` is no longer relay configuration.

#### Scenario: Missing required configuration
- **WHEN** the relay starts without `KAFKA_BROKERS`
- **THEN** startup fails and the relay reports the missing configuration

#### Scenario: Default listen address
- **WHEN** the relay starts without `WS_ADDR`
- **THEN** it listens on `:9000`

#### Scenario: Invalid listen address
- **WHEN** the relay starts with an unusable listen address
- **THEN** startup fails and the relay reports the invalid address

## ADDED Requirements

### Requirement: Relay hosts a WebSocket ingest endpoint
The relay SHALL host a WebSocket server at the fixed path `/telemetry` and accept incoming telemetry records from connected devices for processing.

#### Scenario: Successful record intake from device
- **WHEN** the relay starts with a valid WebSocket listen configuration
- **THEN** it listens for device connections on `/telemetry`
- **THEN** it accepts incoming telemetry records for processing

#### Scenario: Non-telemetry path
- **WHEN** a client attempts a WebSocket handshake on a path other than `/telemetry`
- **THEN** the relay does not upgrade the connection

### Requirement: Relay remains available when device connections close
The relay SHALL continue listening and accept new device connections after any connection closes or fails.

#### Scenario: Device disconnects
- **WHEN** a device connection closes or fails
- **THEN** the relay remains listening for new connections

#### Scenario: New connection after disconnect
- **WHEN** a device reconnects after a disconnect
- **THEN** the relay accepts the new connection and processes its telemetry

## REMOVED Requirements

### Requirement: Relay ingests telemetry from a configured upstream WebSocket
**Reason**: The previous requirement described a Kafka consumer session, but the relay does not consume from Kafka. The relay now hosts a WebSocket ingest endpoint and publishes received records to Kafka.
**Migration**: Devices connect to `ws://<host>:9000/telemetry`; the relay publishes records to Kafka and telemetry-server consumes them from the configured topic.

### Requirement: Relay restores upstream connectivity after instability
**Reason**: The relay no longer dials an upstream source; it hosts the WebSocket server, so outbound reconnection logic does not apply.
**Migration**: Devices reconnect to `ws://<host>:9000/telemetry`; the relay keeps listening and accepts new connections.
