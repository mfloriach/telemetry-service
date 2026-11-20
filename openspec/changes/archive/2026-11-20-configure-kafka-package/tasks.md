# Tasks

## 1. Shared option types and validation

- [x] 1.1 Create `pkg/kafka/options.go` with the package-owned `Acks` and `StartOffset` types, their constants (`AcksNone`, `AcksLeader`, `AcksAll`; `StartOffsetDefault`, `StartOffsetFirst`, `StartOffsetLast`), and the shared positive-magnitude validation helper; verify `go build ./pkg/kafka` succeeds and `go doc server/pkg/kafka.Acks` lists exactly the three acknowledgement constants.
- [x] 1.2 Add table-driven tests for every accepted and rejected enum value and for positive, zero, and negative magnitudes; verify `go test -race ./pkg/kafka` passes with the new validation cases.

## 2. Publisher options

- [x] 2.1 Add `PublisherOption`, `publisherConfig`, and `WithRequiredAcks`, `WithMaxAttempts`, `WithBatchSize`, `WithBatchTimeout`; extend `NewPublisher` to keep the required broker/topic checks first, then apply options and build the `kafka.Writer` with `LeastBytes`; verify `go build ./pkg/kafka` succeeds and `go doc server/pkg/kafka.NewPublisher` shows the variadic `PublisherOption` argument.
- [x] 2.2 Add table-driven publisher tests asserting each option reaches the writer fields, omitted options leave `LeastBytes` and zero-valued tuning intact, and empty brokers or a blank topic return an error naming the input; verify `go test -race ./pkg/kafka` passes.

## 3. Consumer options and call site

- [x] 3.1 Add `ConsumerOption`, `consumerConfig`, and `WithMinBytes`, `WithMaxBytes`, `WithMaxWait`, `WithStartOffset`; change `NewConsumer` to return `(*Consumer, error)`, validate brokers, topic, and group before applying options, and construct the reader only after validation; verify `go build ./pkg/kafka` succeeds and `go doc server/pkg/kafka.NewConsumer` reports the error return.
- [x] 3.2 Add table-driven consumer tests asserting option values through `Reader.Config()`, omitted options leave the library defaults in place, blank required values return naming errors, and non-positive values or an unsupported start offset return an error without panicking; verify `go test -race ./pkg/kafka` passes.
- [x] 3.3 Update `cmd/telemetry-server/main.go` to check the new consumer error and wrap it as `create kafka consumer: %w`; verify `go build ./cmd/...` succeeds and `rg -n "kafka.NewConsumer" --glob '*.go'` shows only the checked assignment.

## 4. Architecture decision record

- [x] 4.1 Present the ADR 0008 draft to the maintainer and obtain approval before writing any file, as required by the change's proposal rules.
- [x] 4.2 Write `docs/adr/0008-configure-kafka-clients-with-functional-options.md` and add its row to `docs/adr/README.md`; verify the linked file exists and the table row matches its title, status, and date.

## 5. Integration verification

- [x] 5.1 Run `gofmt -l pkg/kafka cmd/telemetry-server` and `golangci-lint run ./pkg/kafka/... ./cmd/...`; verify both report no findings.
- [x] 5.2 Run `go test -race ./...` and `go build ./cmd/...`; verify the relay and ingest packages still pass, confirming no message-format or loop behavior changed.

## Workflow follow-up

- Propose a Conventional Commits message for the implementation and validate it with commitlint; do not create the commit unless explicitly requested.
- Archive the change with `/opsx-archive` once implementation and the ADR are complete.
