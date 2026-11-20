# Proposal

## Why

The Kafka adapters expose only the required broker, topic, and group values. Every optional setting is fixed inside the constructors: the publisher always uses `LeastBytes` and Kafka's fire-and-forget `RequiredAcks` default, and the consumer cannot tune fetch size or wait. The relay therefore cannot choose durable acknowledgements or adapt batching and fetch behavior to its load.

## What Changes

- Add functional options to `kafka.NewPublisher`, keeping `brokers` and `topic` as required positional arguments. Existing calls keep working.
- Add functional options to `kafka.NewConsumer`, keeping `brokers`, `topic`, and `groupID` as required positional arguments.
- **BREAKING**: change `NewConsumer` to return `(*Consumer, error)` so invalid optional values fail at construction instead of panicking inside `kafka-go`. Update its single call site in `cmd/telemetry-server/main.go`.
- Publisher options: required acknowledgements, maximum attempts, batch size, and batch timeout.
- Consumer options: minimum and maximum fetch bytes, maximum wait, and start offset.
- Preserve current defaults when an option is omitted: `LeastBytes` for the publisher and the `kafka-go` defaults for everything else.
- Validate option values in the constructors and return errors naming the invalid setting.
- Record the decision in ADR 0008.

## Capabilities

### New Capabilities
- `telemetry/kafka-client`: how the Kafka publisher and consumer are constructed, which parameters are required, which settings are optional, their defaults, and how invalid configuration is reported.

### Modified Capabilities

None.

## Impact

- `pkg/kafka/publisher.go` and `pkg/kafka/consumer.go`, plus one package-local options file.
- One call site: `cmd/telemetry-server/main.go`.
- No new dependencies, no Kafka message-format change, and no change to the relay loop or ingest behavior.
