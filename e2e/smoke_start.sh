#!/usr/bin/env bash
# Verifies that the telemetry server starts, serves HTTP, and shuts down cleanly.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
work="$(mktemp -d)"
addr="127.0.0.1:18080"
body="${work}/body.json"
log="${work}/telemetry-server.log"

pid=""
cleanup() {
  if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
    kill -TERM "$pid" 2>/dev/null || true
    wait "$pid" 2>/dev/null || true
  fi
  rm -rf "$work"
}
trap cleanup EXIT

cd "$root"
go build -o "${work}/telemetry-server" ./cmd/telemetry-server

# Kafka is not required to serve HTTP: the consumer only connects when it reads.
KAFKA_BROKERS="localhost:9092" \
HTTP_ADDR="$addr" \
APP_ENV="production" \
LOG_LEVEL="info" \
"${work}/telemetry-server" >"$log" 2>&1 &
pid=$!

ready="false"
for _ in $(seq 1 50); do
  if ! kill -0 "$pid" 2>/dev/null; then
    echo "telemetry-server exited before becoming ready" >&2
    cat "$log" >&2
    exit 1
  fi
  if curl -fsS "http://${addr}/telemetry/latest?n=1" >"$body" 2>/dev/null; then
    ready="true"
    break
  fi
  sleep 0.2
done

if [ "$ready" != "true" ]; then
  echo "telemetry-server did not serve /telemetry/latest within 10s" >&2
  cat "$log" >&2
  exit 1
fi

if [ "$(cat "$body")" != "[]" ]; then
  echo "unexpected response body: $(cat "$body")" >&2
  exit 1
fi

kill -TERM "$pid"
for _ in $(seq 1 50); do
  if ! kill -0 "$pid" 2>/dev/null; then
    break
  fi
  sleep 0.2
done

if kill -0 "$pid" 2>/dev/null; then
  echo "telemetry-server did not shut down within 10s" >&2
  cat "$log" >&2
  exit 1
fi

if ! wait "$pid"; then
  echo "telemetry-server exited non-zero" >&2
  cat "$log" >&2
  exit 1
fi
pid=""

echo "telemetry-server started, served /telemetry/latest, and shut down cleanly"
