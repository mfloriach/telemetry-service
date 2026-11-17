# Spec Delta

## Purpose
Lets users interact with the WebSocket ingestion system using a standardized interface.

## ADDED Requirements

### Requirement: Connection establishment
The system SHALL establish a WebSocket connection when the ingest client initiates.
- **WHEN** the client calls `Dial(ctx, url)`
- **THEN** a WebSocket connection is established and `Conn` is returned

### Requirement: Message receiving
The system SHALL receive and process incoming WebSocket messages.
- **WHEN** a text binary message is received on the WebSocket
- **THEN** the message is parsed and added to the ingest buffer

### Requirement: Clean disconnection
The system SHALL cleanly disconnect the WebSocket when the session ends.
- **WHEN** the client calls `Close(code, reason)` on the connection
- **THEN** the WebSocket connection is closed gracefully and resources are released
