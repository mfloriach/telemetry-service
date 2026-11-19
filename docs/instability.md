# Instability Handling Notes

This service is designed for an upstream telemetry WebSocket source that may be unstable (periodic disconnects, failed connects, or drops every ~10 seconds for a few seconds).

## Expectations

- **Best-effort ingest:** the service buffers telemetry records it successfully receives; it does not guarantee gap-free or exactly-once delivery from upstream.
- **Automatic reconnect:** on dial/read/connection failures the ingestion loop retries with exponential backoff and jitter (capped), and resets backoff after a healthy successful reconnect/session.
- **Structured logs:** connection, disconnect, and reconnect-scheduling events are logged to aid diagnosis during unstable periods.
- **In-memory buffering only:** the ring buffer retains the latest 1000 records in process memory and does not persist data across restarts. After a restart, buffered telemetry starts empty until new records are ingested.

These behaviors align with the `telemetry` spec scenarios for resilient reconnection and in-memory buffering. Related focused tests live in `internal/ingest` (reconnect/backoff/reset) and `internal/buffer` (capacity, overwrite, no persistence across restart). 
