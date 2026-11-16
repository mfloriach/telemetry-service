// Package buffer provides an in-memory, fixed-capacity ring buffer for satellite
// telemetry records.
//
// Records are stored as the original JSON payloads received from upstream and
// are never persisted across process restarts. The buffer retains at most 1000
// records by default (latest records win when full).
//
// Query results are derived from the most recently ingested records and sorted
// by payload "ts" in descending order. If "ts" is missing or not a comparable
// JSON number, the record is assigned math.MinInt64 for ordering so it sorts
// after records with valid comparable timestamps.
package buffer

import (
	"bytes"
	"encoding/json"
	"math"
	"sort"
	"sync"
)

// DefaultCapacity is the default number of telemetry records retained in memory.
const DefaultCapacity = 1000

// Record stores an ingested telemetry payload and its sortable timestamp key.
type Record struct {
	Payload json.RawMessage
	TS      int64
}

// RingBuffer is a fixed-capacity, concurrency-safe buffer of telemetry records.
// It retains only the most recent records and overwrites the oldest when full.
type RingBuffer struct {
	mu       sync.RWMutex
	capacity int
	records  []Record
	next     int
	size     int
}

// NewRingBuffer creates a ring buffer with the given capacity.
// If capacity <= 0, DefaultCapacity is used.
func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &RingBuffer{
		capacity: capacity,
		records:  make([]Record, capacity),
	}
}

// Add appends a telemetry record payload, extracting its ts sort key.
func (b *RingBuffer) Add(payload []byte) {
	rec := Record{
		Payload: append(json.RawMessage(nil), bytes.TrimSpace(payload)...),
		TS:      extractTS(payload),
	}

	b.mu.Lock()
	b.records[b.next] = rec
	b.next = (b.next + 1) % b.capacity
	if b.size < b.capacity {
		b.size++
	}
	b.mu.Unlock()
}

// Len returns the number of buffered records.
func (b *RingBuffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// Snapshot returns buffered records in ingestion order from oldest to newest.
func (b *RingBuffer) Snapshot() []Record {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]Record, 0, b.size)
	if b.size == 0 {
		return out
	}
	start := 0
	if b.size == b.capacity {
		start = b.next
	}
	for i := 0; i < b.size; i++ {
		idx := (start + i) % b.capacity
		rec := b.records[idx]
		out = append(out, Record{
			Payload: append(json.RawMessage(nil), rec.Payload...),
			TS:      rec.TS,
		})
	}
	return out
}

// Latest returns up to n most recently ingested records sorted by payload ts descending.
// If n <= 0 or no records are buffered, it returns an empty slice.
// Records with missing or uncomparable ts values are assigned math.MinInt64 and
// therefore sort after records with valid comparable ts values.
func (b *RingBuffer) Latest(n int) []Record {
	if n <= 0 {
		return []Record{}
	}

	snap := b.Snapshot()
	if len(snap) == 0 {
		return []Record{}
	}
	if n > len(snap) {
		n = len(snap)
	}

	latest := make([]Record, 0, n)
	for i := len(snap) - n; i < len(snap); i++ {
		latest = append(latest, snap[i])
	}

	sort.SliceStable(latest, func(i, j int) bool {
		return latest[i].TS > latest[j].TS
	})
	return latest
}

func extractTS(payload []byte) int64 {
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
