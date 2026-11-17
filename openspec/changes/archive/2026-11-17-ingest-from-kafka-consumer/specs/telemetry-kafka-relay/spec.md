# Spec Delta

## MODIFIED Requirements

### Requirement: Relay ingests telemetry from a configured upstream WebSocket

The relay SHALL connect to a Kafka consumer group and read incoming telemetry records for processing.

#### Scenario: Successful record intake from Kafka consumer

- **WHEN** the relay starts with a valid Kafka consumer configuration
- **THEN** it establishes a consumer group session and accepts incoming telemetry records for processing

### Requirement: Relay restores upstream connectivity after instability

The relay SHALL rely on the Kafka consumer group's built-in rebalancing and recovery mechanisms to resume consumption after temporary failures, without requiring explicit WebSocket reconnection logic.

#### Scenario: Consumer group rebalance during operation

- **WHEN** a Kafka consumer group rebalance occurs
- **THEN** the relay pauses consumption during rebalance and resumes reading records after partition assignment is restored