# Spec Delta

## Purpose

Provide reliable access to recent satellite telemetry by ingesting an unstable upstream WebSocket feed, buffering recent records, and serving them through an HTTP query endpoint.

## ADDED Requirements

### Requirement: Ingest telemetry from a configured upstream WebSocket
The system SHALL connect to a single configured upstream WebSocket telemetry source and ingest each received JSON telemetry record for buffering.

#### Scenario: Successful ingestion from configured source
- **WHEN** the service starts with a valid upstream WebSocket URL configuration
- **THEN** the service connects to that source and accepts incoming JSON telemetry records for storage

### Requirement: Reconnect automatically during upstream instability
The system SHALL attempt to restore upstream connectivity automatically when the WebSocket connection cannot be established or is lost, using exponential backoff with jitter and resetting backoff after a successful reconnect.

#### Scenario: Connection drops during an instability window
- **WHEN** an established upstream WebSocket connection is lost
- **THEN** the service logs the disconnect and retries connection attempts with exponential backoff and jitter until it reconnects

#### Scenario: Reconnect succeeds after failures
- **WHEN** a retry attempt successfully re-establishes the upstream connection
- **THEN** the service resumes ingesting telemetry records and resets the reconnect backoff state

### Requirement: Buffer recent telemetry in memory
The system SHALL retain only the most recent 1000 ingested telemetry records in an in-memory ring buffer, discarding older records as new records arrive, without persisting buffered data across process restarts.

#### Scenario: Buffer reaches capacity
- **WHEN** more than 1000 telemetry records have been ingested
- **THEN** the buffer contains only the latest 1000 records and older records are no longer available

#### Scenario: Service restart clears buffered data
- **WHEN** the service process restarts
- **THEN** previously buffered in-memory telemetry records are not retained

### Requirement: Serve latest telemetry records over HTTP
The system SHALL expose `GET /telemetry/latest` and return a JSON array of the latest `n` buffered telemetry records sorted in descending order by payload `ts`, where each returned record preserves the original JSON object fields as received.

#### Scenario: Valid request returns latest records
- **WHEN** a client requests `GET /telemetry/latest?n=5` and at least 5 records are buffered
- **THEN** the response is HTTP 200 with a JSON array of 5 records ordered by descending payload `ts`, each matching the original ingested record content

#### Scenario: Fewer records available than requested
- **WHEN** a client requests a valid `n` greater than the number of buffered records
- **THEN** the response is HTTP 200 with a JSON array containing all available buffered records ordered by descending payload `ts`

### Requirement: Validate the telemetry query parameter
The system SHALL require query parameter `n` to be an integer between 1 and 100 inclusive and SHALL reject invalid requests with HTTP 400.

#### Scenario: Missing or non-integer n
- **WHEN** a client requests `GET /telemetry/latest` without a valid integer `n`
- **THEN** the response is HTTP 400

#### Scenario: Out-of-range n
- **WHEN** a client requests `GET /telemetry/latest` with `n` less than 1 or greater than 100
- **THEN** the response is HTTP 400
