package buffer

import (
	"bytes"
	"encoding/json"
	"math"
)

// TelemetryBuffer stores raw telemetry payloads and orders them by their JSON
// "ts" field. Payloads are trimmed and copied on insertion.
type TelemetryBuffer struct {
	buf *Buffer[json.RawMessage]
}

// NewTelemetryBuffer creates a telemetry buffer with the given capacity.
// If capacity <= 0, DefaultCapacity is used.
func NewTelemetryBuffer(capacity int) *TelemetryBuffer {
	return &TelemetryBuffer{
		buf: New(capacity, WithTimestamp(extractTS)),
	}
}

// Add appends a telemetry payload, extracting its ts sort key.
func (b *TelemetryBuffer) Add(payload []byte) {
	trimmed := bytes.TrimSpace(payload)
	value := make(json.RawMessage, len(trimmed))
	copy(value, trimmed)
	b.buf.Add(value)
}

// Len returns the number of buffered payloads.
func (b *TelemetryBuffer) Len() int {
	return b.buf.Len()
}

// Snapshot returns buffered payloads in ingestion order from oldest to newest.
func (b *TelemetryBuffer) Snapshot() []json.RawMessage {
	return clonePayloads(b.buf.Snapshot())
}

// Latest returns up to n most recently ingested payloads sorted by payload ts
// descending. If n <= 0 or no payloads are buffered, it returns an empty slice.
// Payloads with missing or uncomparable ts values are assigned math.MinInt64
// and therefore sort after payloads with valid comparable ts values.
func (b *TelemetryBuffer) Latest(n int) []json.RawMessage {
	return clonePayloads(b.buf.Latest(n))
}

func clonePayloads(values []json.RawMessage) []json.RawMessage {
	out := make([]json.RawMessage, len(values))
	for i, value := range values {
		out[i] = append(json.RawMessage(nil), value...)
	}
	return out
}

func extractTS(payload json.RawMessage) int64 {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(payload, &obj); err != nil {
		return math.MinInt64
	}
	raw, ok := obj["ts"]
	if !ok {
		return math.MinInt64
	}
	return parseTSValue(raw)
}

func parseTSValue(raw json.RawMessage) int64 {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return math.MinInt64
	}

	var i int64
	if err := json.Unmarshal(raw, &i); err == nil {
		return i
	}

	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return math.MinInt64
		}
		return int64(f)
	}

	return math.MinInt64
}
