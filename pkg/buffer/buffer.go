// Package buffer provides a concurrency-safe, fixed-capacity ring buffer.
//
// The buffer stores values of a caller-chosen type and retains at most a fixed
// number of them. It is in-memory only; values are lost when the process exits.
//
// Latest orders retained values by a timestamp extracted from each value. When
// no extractor is configured, values keep their ingestion order.
package buffer

import (
	"sort"
	"sync"
)

// DefaultCapacity is the default number of values retained in memory.
const DefaultCapacity = 1000

// Option configures a Buffer.
type Option[T any] func(*config[T])

type config[T any] struct {
	timestamp func(T) int64
}

// WithTimestamp sets the function used to extract the ordering key from a
// value. Latest sorts larger timestamps first. When omitted, Latest preserves
// ingestion order.
func WithTimestamp[T any](fn func(T) int64) Option[T] {
	return func(cfg *config[T]) {
		cfg.timestamp = fn
	}
}

type entry[T any] struct {
	value T
	ts    int64
}

// Buffer is a fixed-capacity, concurrency-safe ring buffer.
// It retains only the most recent values and overwrites the oldest when full.
type Buffer[T any] struct {
	mu        sync.RWMutex
	capacity  int
	entries   []entry[T]
	next      int
	size      int
	timestamp func(T) int64
}

// New creates a Buffer with the given capacity.
// If capacity <= 0, DefaultCapacity is used.
func New[T any](capacity int, opts ...Option[T]) *Buffer[T] {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}

	cfg := config[T]{}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return &Buffer[T]{
		capacity:  capacity,
		entries:   make([]entry[T], capacity),
		timestamp: cfg.timestamp,
	}
}

// Add appends a value, overwriting the oldest when the buffer is full.
func (b *Buffer[T]) Add(value T) {
	var ts int64
	if b.timestamp != nil {
		ts = b.timestamp(value)
	}

	b.mu.Lock()
	b.entries[b.next] = entry[T]{value: value, ts: ts}
	b.next = (b.next + 1) % b.capacity
	if b.size < b.capacity {
		b.size++
	}
	b.mu.Unlock()
}

// Len returns the number of buffered values.
func (b *Buffer[T]) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// Snapshot returns buffered values in ingestion order from oldest to newest.
func (b *Buffer[T]) Snapshot() []T {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]T, 0, b.size)
	if b.size == 0 {
		return out
	}

	start := 0
	if b.size == b.capacity {
		start = b.next
	}
	for i := 0; i < b.size; i++ {
		out = append(out, b.entries[(start+i)%b.capacity].value)
	}
	return out
}

// Latest returns up to n most recently ingested values sorted by timestamp
// descending. If n <= 0 or no values are buffered, it returns an empty slice.
// Values with equal timestamps keep their ingestion order.
func (b *Buffer[T]) Latest(n int) []T {
	if n <= 0 {
		return []T{}
	}

	b.mu.RLock()
	if b.size == 0 {
		b.mu.RUnlock()
		return []T{}
	}
	if n > b.size {
		n = b.size
	}

	start := (b.next - n + b.capacity) % b.capacity
	latest := make([]entry[T], n)
	for i := 0; i < n; i++ {
		latest[i] = b.entries[(start+i)%b.capacity]
	}
	b.mu.RUnlock()

	if b.timestamp != nil {
		sort.SliceStable(latest, func(i, j int) bool {
			return latest[i].ts > latest[j].ts
		})
	}

	out := make([]T, n)
	for i, e := range latest {
		out[i] = e.value
	}
	return out
}
