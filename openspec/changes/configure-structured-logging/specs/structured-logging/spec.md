# Spec Delta

## Purpose

Define configurable, environment-aware structured logging with context-correlated records for the telemetry server.

## ADDED Requirements

### Requirement: Environment-aware log format selection
The system SHALL select the log output format based on runtime configuration. When `LOG_FORMAT` is not explicitly set, the system SHALL use JSON format when `APP_ENV` is `production` (case-insensitive) and text format for all other `APP_ENV` values. When `LOG_FORMAT` is explicitly set to `json` or `text` (case-insensitive), the system SHALL use the configured format.

#### Scenario: Production environment defaults to JSON logs
- **WHEN** the service starts with `APP_ENV=production` and no explicit `LOG_FORMAT`
- **THEN** application logs are emitted in JSON format

#### Scenario: Non-production environment defaults to text logs
- **WHEN** the service starts with `APP_ENV=development` or `APP_ENV` unset and no explicit `LOG_FORMAT`
- **THEN** application logs are emitted in text format

#### Scenario: Explicit log format overrides environment default
- **WHEN** the service starts with `APP_ENV=production` and `LOG_FORMAT=text`
- **THEN** application logs are emitted in text format

### Requirement: Configurable log level
The system SHALL configure the application log level from runtime configuration. `LOG_LEVEL` MUST accept `debug`, `info`, `warn`, and `error` (case-insensitive), default to `info` when unset, and apply the configured level to emitted application logs.

#### Scenario: Default log level is info
- **WHEN** the service starts without an explicit `LOG_LEVEL`
- **THEN** the logger is configured at info level

#### Scenario: Explicit log level is applied
- **WHEN** the service starts with `LOG_LEVEL=debug`
- **THEN** debug-level and higher application logs are eligible for emission according to the configured logger

### Requirement: Logging configuration sources and validation
The system SHALL read logging configuration from environment variables and CLI flags. CLI flags `-log-level` and `-log-format` MUST take precedence over `LOG_LEVEL` and `LOG_FORMAT` environment variables. The system MUST fail fast at startup when `APP_ENV`-derived/defaulted or explicit `LOG_FORMAT` is not `json` or `text`, or when `LOG_LEVEL` is not a supported value.

#### Scenario: CLI flag overrides environment variable
- **WHEN** the service starts with `LOG_LEVEL=info` and `-log-level=error`
- **THEN** the logger is configured at error level

#### Scenario: Invalid logging configuration is rejected
- **WHEN** the service starts with an unsupported `LOG_LEVEL` or `LOG_FORMAT` value
- **THEN** startup fails with a configuration error before the service begins serving traffic

### Requirement: Context-aware application logging
The system SHALL emit application logs with an associated context. Application log sites MUST use context-aware logging so that request and runtime context can be attached to log records. When a `request_id` is present in the context, emitted logs for that context MUST include the request correlation identifier.

#### Scenario: Request-scoped logs include request correlation
- **WHEN** a request with an assigned request ID is processed and the application emits a log for that request context
- **THEN** the log record includes the request ID

#### Scenario: Background logs are emitted with runtime context
- **WHEN** the service emits logs during startup, shutdown, or upstream ingestion processing
- **THEN** the logs are emitted through the configured application logger with the relevant runtime context
