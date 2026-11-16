package buffer

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

func TestRingBuffer_AddAppendAndOverwrite(t *testing.T) {
	t.Parallel()

	b := NewRingBuffer(3)
	for i := 1; i <= 5; i++ {
		b.Add([]byte(fmt.Sprintf(`{"satellite":"AE1","ts":%d,"battery":%d}`, i, i)))
	}

	if got := b.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}

	snap := b.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("Snapshot() len = %d, want 3", len(snap))
	}

	wantTS := []int64{3, 4, 5}
	for i, rec := range snap {
		if rec.TS != wantTS[i] {
			t.Fatalf("Snapshot()[%d].TS = %d, want %d", i, rec.TS, wantTS[i])
		}
		var obj map[string]any
		if err := json.Unmarshal(rec.Payload, &obj); err != nil {
			t.Fatalf("Snapshot()[%d] payload unmarshal: %v", i, err)
		}
		if got := int64(obj["ts"].(float64)); got != wantTS[i] {
			t.Fatalf("Snapshot()[%d] payload ts = %d, want %d", i, got, wantTS[i])
		}
	}
}

func TestRingBuffer_OverwriteKeepsLatest1000(t *testing.T) {
	t.Parallel()

	b := NewRingBuffer(DefaultCapacity)
	total := 1205
	for i := 1; i <= total; i++ {
		b.Add([]byte(fmt.Sprintf(`{"satellite":"AE1","ts":%d}`, i)))
	}

	if got := b.Len(); got != DefaultCapacity {
		t.Fatalf("Len() = %d, want %d", got, DefaultCapacity)
	}

	snap := b.Snapshot()
	if len(snap) != DefaultCapacity {
		t.Fatalf("Snapshot() len = %d, want %d", len(snap), DefaultCapacity)
	}

	firstWant := int64(total - DefaultCapacity + 1)
	if snap[0].TS != firstWant {
		t.Fatalf("oldest TS = %d, want %d", snap[0].TS, firstWant)
	}
	if snap[len(snap)-1].TS != int64(total) {
		t.Fatalf("newest TS = %d, want %d", snap[len(snap)-1].TS, total)
	}
}

func TestRingBuffer_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	b := NewRingBuffer(100)
	const writers = 5
	const perWriter = 200

	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				ts := w*perWriter + i
				b.Add([]byte(fmt.Sprintf(`{"satellite":"AE1","ts":%d}`, ts)))
			}
		}()
	}

	done := make(chan struct{})
	var readers sync.WaitGroup
	for r := 0; r < 4; r++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-done:
					return
				default:
					snap := b.Snapshot()
					if l := b.Len(); l < 0 || l > 100 {
						t.Errorf("Len() out of range: %d", l)
						return
					}
					if len(snap) > 100 {
						t.Errorf("Snapshot() len out of range: %d", len(snap))
						return
					}
				}
			}
		}()
	}

	wg.Wait()
	close(done)
	readers.Wait()

	if got := b.Len(); got != 100 {
		t.Fatalf("Len() after concurrent writes = %d, want 100", got)
	}
	if got := len(b.Snapshot()); got != 100 {
		t.Fatalf("Snapshot() len after concurrent writes = %d, want 100", got)
	}
}

func TestExtractTS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload string
		want    int64
	}{
		{name: "integer", payload: `{"ts":1735657200400}`, want: 1735657200400},
		{name: "float", payload: `{"ts":1735657200400.9}`, want: 1735657200400},
		{name: "missing", payload: `{"battery":1}`, want: -9223372036854775808},
		{name: "invalid json", payload: `{`, want: -9223372036854775808},
		{name: "non numeric", payload: `{"ts":"x"}`, want: -9223372036854775808},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := extractTS([]byte(tt.payload)); got != tt.want {
				t.Fatalf("extractTS(%s) = %d, want %d", tt.payload, got, tt.want)
			}
		})
	}
}

func TestRingBuffer_LatestSortedByTSDesc(t *testing.T) {
	t.Parallel()

	b := NewRingBuffer(10)
	b.Add([]byte(`{"satellite":"AE1","ts":100,"battery":1}`))
	b.Add([]byte(`{"satellite":"AE1","ts":300,"battery":3}`))
	b.Add([]byte(`{"satellite":"AE1","ts":200,"battery":2}`))

	got := b.Latest(3)
	if len(got) != 3 {
		t.Fatalf("Latest(3) len = %d, want 3", len(got))
	}

	wantTS := []int64{300, 200, 100}
	for i, rec := range got {
		if rec.TS != wantTS[i] {
			t.Fatalf("Latest(3)[%d].TS = %d, want %d", i, rec.TS, wantTS[i])
		}
		var obj map[string]any
		if err := json.Unmarshal(rec.Payload, &obj); err != nil {
			t.Fatalf("Latest(3)[%d] unmarshal: %v", i, err)
		}
		if obj["battery"].(float64) != float64(wantTS[i]/100) {
			t.Fatalf("Latest(3)[%d] payload not preserved as expected: %s", i, string(rec.Payload))
		}
		if _, ok := obj["satellite"]; !ok {
			t.Fatalf("Latest(3)[%d] missing original field satellite", i)
		}
	}
}

func TestRingBuffer_LatestFewerThanRequested(t *testing.T) {
	t.Parallel()

	b := NewRingBuffer(10)
	b.Add([]byte(`{"satellite":"AE1","ts":10}`))
	b.Add([]byte(`{"satellite":"AE1","ts":20}`))

	got := b.Latest(5)
	if len(got) != 2 {
		t.Fatalf("Latest(5) len = %d, want 2", len(got))
	}
	if got[0].TS != 20 || got[1].TS != 10 {
		t.Fatalf("Latest(5) order = [%d %d], want [20 10]", got[0].TS, got[1].TS)
	}
}

func TestRingBuffer_LatestMissingTSSortsLast(t *testing.T) {
	t.Parallel()

	b := NewRingBuffer(10)
	b.Add([]byte(`{"satellite":"AE1","battery":1}`))
	b.Add([]byte(`{"satellite":"AE1","ts":50}`))
	b.Add([]byte(`{"satellite":"AE1","ts":"bad"}`))

	got := b.Latest(3)
	if len(got) != 3 {
		t.Fatalf("Latest(3) len = %d, want 3", len(got))
	}
	if got[0].TS != 50 {
		t.Fatalf("Latest(3)[0].TS = %d, want 50", got[0].TS)
	}
	if got[1].TS != -9223372036854775808 || got[2].TS != -9223372036854775808 {
		t.Fatalf("missing/uncomparable ts should sort last, got [%d %d %d]", got[0].TS, got[1].TS, got[2].TS)
	}
	if string(got[0].Payload) != `{"satellite":"AE1","ts":50}` {
		t.Fatalf("payload pass-through mismatch: %s", string(got[0].Payload))
	}
}

// TestRingBuffer_NoPersistenceAcrossRestart documents the in-memory-only expectation:
// a newly created buffer does not retain records from a previous process/instance.
func TestRingBuffer_NoPersistenceAcrossRestart(t *testing.T) {
	t.Parallel()

	prev := NewRingBuffer(10)
	prev.Add([]byte(`{"satellite":"AE1","ts":1}`))
	if prev.Len() != 1 {
		t.Fatalf("pre-restart Len() = %d, want 1", prev.Len())
	}

	restarted := NewRingBuffer(10)
	if restarted.Len() != 0 {
		t.Fatalf("post-restart Len() = %d, want 0", restarted.Len())
	}
	if got := restarted.Latest(5); len(got) != 0 {
		t.Fatalf("post-restart Latest(5) len = %d, want 0", len(got))
	}
}
