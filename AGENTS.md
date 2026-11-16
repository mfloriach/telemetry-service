# AGENTS.md

## Project Overview
- **Language:** Go 1.27+
- **Architecture:** Standard Go project layout (`cmd/` for binaries, `internal/` for private domain logic).
- **Database:** PostgreSQL using `pgx/v5` and `sqlc` for generated queries.

## Build, Test & Lint Commands
Run these commands to verify changes before finishing a task:
- **Format:** `gofmt -w .`
- **Build:** `go build ./cmd/...`
- **Test (single package):** `go test -v -race ./internal/path/to/pkg`
- **Test (all):** `go test -race ./...`
- **Lint:** `golangci-lint run ./...`

## Go Code Style & Idioms
- **Simplicity:** Favor standard library packages over third-party dependencies.
- **Error Handling:** Always check errors immediately. Wrap errors with context using `fmt.Errorf("action failed: %w", err)`. Never ignore errors with `_` unless explicitly justified.
- **Control Flow:** Use early returns to keep the happy path left-aligned; avoid deep `if-else` nesting.
- **Naming:** Use short, single-word, lowercase package names. Never create `util`, `common`, or `base` packages.
- **Context:** Always pass `ctx context.Context` as the first parameter to functions performing I/O or network calls.

## Testing Guidelines
- Write table-driven tests for functions with multiple input/output cases.
- Use `t.Helper()` in test helper functions and `t.Parallel()` where safe.
- Place unit tests in `*_test.go` files alongside the source code when is possible and inside folder `e2e`.

## Boundaries & Constraints
- **Always:** Run `go test -race` and `golangci-lint run` on modified packages before completing a task.
- **Ask First:** Before adding new external dependencies to `go.mod`, changing public API contracts, or writing new SQL migrations.
- **Never:** Manually edit generated files (`*.sql.go`, `*.pb.go`, or `mock_*.go`). Regenerate them via `make generate` instead.
