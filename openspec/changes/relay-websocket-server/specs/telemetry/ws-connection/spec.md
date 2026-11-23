# Spec Delta

## ADDED Requirements

### Requirement: Server listens at a fixed telemetry path
The server SHALL listen for WebSocket handshakes on the configured address at the fixed path `/telemetry`.

#### Scenario: Valid telemetry path
- **WHEN** the server is started with a valid listen address
- **THEN** it accepts a WebSocket handshake on `/telemetry`

#### Scenario: Non-telemetry path
- **WHEN** a client attempts a WebSocket handshake on a path other than `/telemetry`
- **THEN** the server does not upgrade the connection

### Requirement: Server configuration is validated before listening
The server SHALL reject a non-positive read limit and an unusable listen address before accepting connections.

#### Scenario: Non-positive read limit
- **WHEN** the server is created with a read limit of zero or less
- **THEN** creation fails with an error naming the read limit
- **THEN** no connection is accepted

#### Scenario: Unusable listen address
- **WHEN** the server cannot use the configured listen address
- **THEN** startup fails with an error naming the address

### Requirement: Server delivers messages intact up to the read limit
The server SHALL deliver every inbound text or binary message that fits within the configured read limit and SHALL reject messages above the limit without returning a partial payload. The default limit is 1 MiB.

#### Scenario: Message within the limit
- **WHEN** a device sends a message smaller than or equal to the read limit
- **THEN** the server delivers the complete payload and its frame type

#### Scenario: Message above the limit
- **WHEN** a device sends a message larger than the read limit
- **THEN** the read fails with an error
- **THEN** no truncated payload is processed

#### Scenario: No limit configured
- **WHEN** the server is created without an explicit read limit
- **THEN** the 1 MiB default applies

### Requirement: Server closes connections with a status code and reason
The server SHALL close each connection with an appropriate WebSocket status code and reason when the client disconnects, when a read fails, or when the server shuts down.

#### Scenario: Client disconnects normally
- **WHEN** a device closes the connection normally
- **THEN** the server releases the connection and continues serving other devices

#### Scenario: Server shuts down
- **WHEN** the server shuts down while a device is connected
- **THEN** the server closes the connection with a going-away status and a reason
- **THEN** the connection is released

#### Scenario: Read failure
- **WHEN** reading from a device connection fails
- **THEN** the server closes the connection with a server-error status and a reason
- **THEN** the failure is recorded

### Requirement: Server operation is observable
The server SHALL record listen failures, accepted connections, and connection close or read errors through the configured logger, and SHALL not log when no logger is configured.

#### Scenario: Listen failure is logged
- **WHEN** the server cannot start listening and a logger is configured
- **THEN** the failure is recorded with the address and the error

#### Scenario: No logger configured
- **WHEN** the server is created without a logger
- **THEN** connection handling proceeds unchanged and nothing is logged
