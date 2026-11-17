# Tasks

## 1. Interface Extraction

- [ ] 1.1 Move `MessageType`, `Message`, `Conn` interface from `internal/ingest/ws.go` to `internal/utils/`
- [ ] 1.2 Move `Dialer` interface from `internal/ingest/ws.go` to `internal/utils/`
- [ ] 1.3 Update `internal/ingest/ingest.go` to import `Dialer` from `internal/utils` instead of using concrete types
- [ ] 1.4 Verify `SessionIngester` compiles with the new `Dialer` interface import

## 2. Retry Mechanism Refactoring

- [ ] 2.1 Refactor `internal/ingest/retry.go` to use `utils.WithRetry` instead of duplicate backoff logic
- [ ] 2.2 Verify `RetryingIngester.Run` maintains identical backoff behavior (100ms base, 5s max, 1s successMin)
- [ ] 2.3 Run existing ingest tests to confirm no regression in retry behavior

## 3. Code Verification

- [ ] 3.1 Run `go build ./...` to confirm all packages compile
- [ ] 3.2 Run `go test -race ./internal/ingest/...` to verify no test failures
- [ ] 3.3 Run `go test -race ./e2e/...` to verify end-to-end functionality
- [ ] 3.4 Check `go vet ./...` for any issues

## 4. Documentation and Examples

- [ ] 4.1 Verify proposal.md, specs/ingest/websocket-interface/spec.md, and design.md are consistent
- [ ] 4.2 Confirm C4 diagram in design.md reflects the new architecture
- [ ] 4.2 Update any internal documentation referencing the moved types

## 5. Workflow Follow-up

- [ ] 5.1 Archive the change with `/opsx-archive` after project review
- [ ] 5.2 Verify archived result and main specs integrity
- [ ] 5.3 Run post-archive validation to ensure all main specs are valid
