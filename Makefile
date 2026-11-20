.PHONY: install-precommit fmt-check lint test build smoke

install-precommit:
	@command -v lefthook >/dev/null 2>&1 || { \
		echo "lefthook is not installed. Install it with: go install github.com/evilmartians/lefthook@latest"; \
		exit 1; \
	}
	lefthook install

fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

lint:
	golangci-lint run ./...

test:
	go test -race ./...

build:
	go build ./cmd/...

smoke:
	./e2e/smoke_start.sh

device:
	for i in $(seq 1 100); do
		echo "{\"device_id\":\"device-001\",\"value\":$i}"
		sleep 1
	done | websocat ws://localhost:8080/ws
