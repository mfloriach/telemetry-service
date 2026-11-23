# Spec Delta

## MODIFIED Requirements

### Requirement: Telemetry records shall be published to Kafka
The system SHALL publish telemetry messages received on each accepted WebSocket connection to a Kafka queue for downstream consumption, preserving the original payload.

#### Scenario: Successful WebSocket to Kafka publish
- **WHEN** a telemetry message is received from a connected device
- **THEN** the message is published to the Kafka queue with the original payload

### Requirement: Kafka publisher shall deliver messages at least once
The system SHALL retry transient Kafka publish failures and SHALL deliver each accepted telemetry message at least once while retries remain.

#### Scenario: Message delivery with retry
- **WHEN** the initial Kafka publish fails due to a transient error
- **THEN** the system retries publishing the message with exponential backoff
- **THEN** the message is delivered to the Kafka queue

#### Scenario: Retries exhausted
- **WHEN** every publish attempt fails
- **THEN** the system stops retrying that message
- **THEN** the affected device connection is closed

## ADDED Requirements

### Requirement: Relay accepts concurrent device connections
The relay SHALL accept more than one device connection at a time and SHALL process each connection independently.

#### Scenario: Multiple devices publish concurrently
- **WHEN** two devices are connected and both send telemetry
- **THEN** both records are published to Kafka independently

#### Scenario: One connection fails
- **WHEN** one device connection fails or closes
- **THEN** other connected devices remain connected and continue publishing

### Requirement: Relay closes the affected connection on publish failure
When a telemetry message cannot be published, the relay SHALL close the affected device connection with a server-error status code and reason, without affecting other connections.

#### Scenario: Publish failure closes only that connection
- **WHEN** publishing a record from one device fails
- **THEN** the relay closes that device connection with a server-error status
- **THEN** other device connections remain open

## REMOVED Requirements

### Requirement: WebSocket connection failures shall not lose telemetry
**Reason**: Server-side buffering while Kafka is unavailable is out of scope for this change; failures are surfaced to the device instead of being silently buffered.
**Migration**: Devices reconnect and resend failed records. Durable buffering remains the responsibility of Kafka and downstream consumers.
