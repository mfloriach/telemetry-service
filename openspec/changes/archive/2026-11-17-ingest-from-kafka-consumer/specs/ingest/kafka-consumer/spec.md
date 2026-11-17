# Spec Delta

## Purpose

Ingest telemetry records from a Kafka consumer into a buffer for downstream processing, decoupling ingest from WebSocket transport concerns.

## ADDED Requirements

### Requirement: Ingest records shall be read from a Kafka consumer

The system SHALL read telemetry records from a Kafka consumer and append each record's payload to the ingest buffer.

#### Scenario: Successful record consumption from Kafka

- **WHEN** a Kafka consumer receives a message containing a telemetry record
- **THEN** the record's payload is appended to the ingest buffer

### Requirement: Kafka consumer context cancellation shall stop ingestion

The system SHALL stop consuming from the Kafka consumer when the context is canceled and return ctx.Err().

#### Scenario: Context canceled during consumption

- **WHEN** the context is canceled while the consumer is reading messages
- **THEN** the ingest loop stops and returns context.Canceled

### Requirement: Kafka consumer read errors shall be surfaced

The system SHALL return a wrapped error when the Kafka consumer fails to read a message, unless the error is a context cancellation or deadline exceeded.

#### Scenario: Consumer read error is reported

- **WHEN** the Kafka consumer returns a read error that is not context-related
- **THEN** the ingest loop returns a wrapped error describing the consumer failure