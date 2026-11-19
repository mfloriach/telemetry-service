# Telemetry WebSocket Connection Specification

## Purpose

Defines how the relay's upstream WebSocket client is configured and behaves: the endpoint it requires, the limits and deadlines it enforces, and what it reports when a connection cannot be established, reads, or closes.

## Requirements

### Requirement: Connecting requires an upstream endpoint

The client SHALL require a non-empty WebSocket endpoint and SHALL fail before opening any connection when it is missing.

#### Scenario: Missing endpoint
- **WHEN** the client is created without an endpoint
- **THEN** it returns an error naming the endpoint as the missing value
- **THEN** no network connection is attempted

#### Scenario: Valid endpoint
- **WHEN** the client is created with a reachable WebSocket endpoint
- **THEN** it completes the opening handshake and is ready to read messages

### Requirement: The opening handshake is bounded by a deadline

The client SHALL abandon the opening handshake once the configured deadline elapses, and SHALL apply a default of 10 seconds when no deadline is configured.

#### Scenario: Slow server
- **WHEN** the endpoint does not complete the handshake within the deadline
- **THEN** the connect call fails with an error identifying a timeout
- **THEN** the partial connection is released

#### Scenario: No deadline configured
- **WHEN** the client is created without an explicit deadline
- **THEN** the 10 second default applies

### Requirement: Messages are delivered intact up to the configured read limit

The client SHALL deliver every inbound message that fits within the configured read limit, together with its frame type, and SHALL report an error instead of a partial payload when a message exceeds it. The default limit is 1 MiB.

#### Scenario: Message within the limit
- **WHEN** an inbound message is smaller than or equal to the read limit
- **THEN** the read returns the complete payload and its frame type

#### Scenario: Message above the limit
- **WHEN** an inbound message exceeds the read limit
- **THEN** the read fails with an error
- **THEN** no truncated payload is returned to the caller

#### Scenario: No limit configured
- **WHEN** the client is created without an explicit read limit
- **THEN** the 1 MiB default applies

### Requirement: Compression is negotiated only when enabled

The client SHALL offer WebSocket compression only when compression is enabled, and SHALL use the configured threshold to decide which messages are compressed. Compression is disabled by default.

#### Scenario: Compression enabled
- **WHEN** compression is enabled and the server accepts it
- **THEN** messages continue to be delivered intact over the compressed connection

#### Scenario: Compression not enabled
- **WHEN** the client is created without enabling compression
- **THEN** no compression extension is offered during the handshake

### Requirement: Handshake headers and subprotocols are configurable

The client SHALL send the configured request headers with the opening handshake, SHALL offer the configured subprotocols, and SHALL adopt the subprotocol the server selects.

#### Scenario: Authentication header
- **WHEN** a request header is configured
- **THEN** it is present in the opening handshake request

#### Scenario: Server selects a subprotocol
- **WHEN** the client offers subprotocols and the server selects one
- **THEN** the connection proceeds with the selected subprotocol

### Requirement: Invalid configuration fails before dialing

The client SHALL reject configuration that cannot produce a valid connection — a non-positive deadline, a non-positive read limit, or a compression threshold supplied while compression is disabled — before any network activity.

#### Scenario: Non-positive read limit
- **WHEN** the read limit is zero or negative
- **THEN** creation fails with an error naming the read limit
- **THEN** no connection is attempted

#### Scenario: Compression threshold without compression
- **WHEN** a compression threshold is supplied but compression is not enabled
- **THEN** creation fails with an error describing the combination

### Requirement: Closing reports the code and reason and releases the connection

The client SHALL send the configured close code and reason when closing and SHALL release the underlying connection, returning any failure to the caller.

#### Scenario: Relay shuts down
- **WHEN** the caller closes the connection with a code and reason
- **THEN** the close frame carrying that code and reason is sent
- **THEN** the underlying connection is released

#### Scenario: Close fails
- **WHEN** the close handshake cannot complete
- **THEN** the failure is returned to the caller
- **THEN** the connection is still released

### Requirement: Connection outcomes are observable

The client SHALL report connection and close failures through the configured logger, and SHALL behave without logging output when no logger is configured.

#### Scenario: Dial failure is logged
- **WHEN** the connection cannot be established and a logger is configured
- **THEN** the failure is recorded with the endpoint and the error

#### Scenario: No logger configured
- **WHEN** the client is created without a logger
- **THEN** connection handling proceeds unchanged and nothing is logged
