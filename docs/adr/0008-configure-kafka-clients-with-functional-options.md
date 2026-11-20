# ADR-0008: Configure the Kafka clients with functional options and explicit required values

**Date**: 2026-11-20
**Status**: accepted
**Deciders**: Project maintainer; coding agent

## Context

`pkg/kafka` exposes only the values a caller must supply and fixes everything else inside its constructors. The publisher always uses `LeastBytes` and Kafka's fire-and-forget acknowledgement default; the consumer cannot tune fetch size or wait. The relay therefore cannot request durable acknowledgements or adapt batching and fetch behavior to its load.

The consumer constructor also cannot report invalid configuration. `kafka-go` panics in `NewReader` when `ReaderConfig` fails validation, so any option that reaches the reader must be validated before construction. There are two call sites in this repository, and ADR-0007 already established the same constructor shape for the WebSocket client.

## Decision

Build Kafka clients through:

- `NewPublisher(brokers []string, topic string, opts ...PublisherOption)`
- `NewConsumer(brokers []string, topic string, groupID string, opts ...ConsumerOption) (*Consumer, error)`

Required connection values stay positional so the compiler enforces them. Optional settings are functional options: acknowledgements, attempts, batch size, and batch timeout for the publisher; minimum/maximum fetch bytes, maximum wait, and start offset for the consumer. The package owns its `Acks` and `StartOffset` types so callers do not import `kafka-go` and cannot supply arbitrary provider integers. Each option validates its argument, and the constructors return an error before creating a writer or reader. Omitted options keep the provider defaults, with `LeastBytes` as the publisher's only package-chosen default.

## Alternatives Considered

### Alternative 1: A configuration struct per client
- **Pros**: Familiar shape; fields are self-documenting.
- **Cons**: The zero value is ambiguous, so "unset" and a meaningful zero need pointers or sentinels.
- **Why not**: It moves defaulting into every caller instead of keeping it in one constructor.

### Alternative 2: Accept `kafka.RequiredAcks` and raw offset integers
- **Pros**: No new option types; direct access to provider semantics.
- **Cons**: Leaks `kafka-go` into the package contract and permits offsets that make `NewReader` panic.
- **Why not**: The package would expose its implementation rather than its own configuration surface.

### Alternative 3: Keep the consumer constructor infallible
- **Pros**: No signature change; existing calls compile unchanged.
- **Cons**: Invalid configuration cannot be reported, and `kafka-go` still panics on a bad `ReaderConfig`.
- **Why not**: The only safe option set would be one that cannot fail, which excludes useful tuning.

## Consequences

### Positive
- One construction path carries every configurable setting, with defaults that match current behavior.
- Invalid configuration surfaces at startup with an error naming the setting, before any network activity.
- Callers keep provider types out of their code, and the relay keeps its existing call unchanged.

### Negative
- `NewConsumer` is now breaking for any out-of-tree caller; the single in-repo call site is updated.
- `WithMinBytes` must be paired with `WithMaxBytes`, because the provider validates the bounds before applying its maximum default.
- The option surface is intentionally narrow: it covers delivery and fetch tuning, not every provider field.

### Risks
- Provider defaults can change across `kafka-go` upgrades — mitigated by tracking them rather than copying them.
- Construction success does not prove broker connectivity — dialing remains lazy, as before.
- Enum types add public surface — limited to three constants each.

## References

- `pkg/kafka/options.go` — option types, validation, and defaults.
- `openspec/changes/configure-kafka-package/design.md` — design decisions for this change.
- `docs/adr/0007-configure-websocket-connection-with-functional-options.md` — the same pattern for the WebSocket client.
