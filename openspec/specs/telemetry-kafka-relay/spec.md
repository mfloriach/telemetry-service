# telemetry-kafka-relay Specification

## Purpose

Define a standalone relay capability that publishes upstream telemetry records to Kafka for downstream consumption without changing existing telemetry-server behavior.

## Requirements

### Requirement: Relay ingests telemetry from a configured upstream WebSocket
The relay SHALL connect to a configured upstream WebSocket telemetry source and read incoming JSON telemetry records for publishing.

#### Scenario: Successful upstream connection and record intake
- **WHEN** the relay starts with valid upstream WebSocket configuration
- **THEN** it establishes a connection to the configured source and accepts incoming JSON telemetry records for publish processing

### Requirement: Relay publishes telemetry records to Kafka
The relay SHALL publish each ingested JSON telemetry record to a configured Kafka broker set and topic as a message whose value preserves the original record content.

#### Scenario: Record received from upstream
- **WHEN** the relay receives a JSON telemetry record from the upstream source
- **THEN** it publishes a Kafka message to the configured topic containing that record's content

### Requirement: Relay validates required runtime configuration
The relay SHALL require configuration for the upstream WebSocket URL, Kafka brokers, and Kafka topic, and SHALL fail startup with a clear error when required configuration is missing or invalid.

#### Scenario: Missing required configuration
- **WHEN** the relay starts without a required upstream URL, broker, or topic value
- **THEN** startup fails and the relay reports which required configuration is missing or invalid

### Requirement: Relay restores upstream connectivity after instability
The relay SHALL attempt to restore upstream WebSocket connectivity automatically after connection failures or disconnects using bounded retry delays, and SHALL resume publishing after reconnection.

#### Scenario: Upstream disconnect during operation
- **WHEN** an established upstream WebSocket connection is lost
- **THEN** the relay retries the upstream connection and resumes reading and publishing telemetry records after reconnecting

### Requirement: Local development environment provides Kafka via Compose
The project SHALL provide a Docker Compose configuration that starts a local single-broker Kafka service in KRaft mode for relay development and testing.

#### Scenario: Developer starts local Kafka environment
- **WHEN** a developer starts the provided Docker Compose configuration
- **THEN** a local Kafka broker is available for the relay to connect to using documented client settings

### Requirement: Relay operation does not alter existing telemetry-server behavior
The introduction of the relay capability SHALL NOT change the existing telemetry-server ingest, buffering, HTTP API, or runtime behavior.

#### Scenario: Existing server behavior remains unchanged
- **WHEN** the relay capability and local Kafka environment are added
- **THEN** existing telemetry-server interfaces and behaviors continue to operate as previously specified
