package buffer

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

func TestBuffer_AddAppendAndOverwrite(t *testing.T) {
	t.Parallel()

	b := New[int](3, WithTimestamp(func(v int) int64 { return int64(v) }))
	for i := 1; i <= 5; i++ {
		b.Add(i)
	}

	if got := b.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}

	snap := b.Snapshot()
	want := []int{3, 4, 5}
	if len(snap) != len(want) {
		t.Fatalf("Snapshot() len = %d, want %d", len(snap), len(want))
	}
	for i, value := range snap {
		if value != want[i] {
			t.Fatalf("Snapshot()[%d] = %d, want %d", i, value, want[i])
		}
	}
}

func TestBuffer_OverwriteKeepsLatestDefaultCapacity(t *testing.T) {
	t.Parallel()

	b := New[int](DefaultCapacity, WithTimestamp(func(v int) int64 { return int64(v) }))
	total := 1205
	for i := 1; i <= total; i++ {
		b.Add(i)
	}

	if got := b.Len(); got != DefaultCapacity {
		t.Fatalf("Len() = %d, want %d", got, DefaultCapacity)
	}

	snap := b.Snapshot()
	if len(snap) != DefaultCapacity {
		t.Fatalf("Snapshot() len = %d, want %d", len(snap), DefaultCapacity)
	}
	if snap[0] != total-DefaultCapacity+1 {
		t.Fatalf("oldest value = %d, want %d", snap[0], total-DefaultCapacity+1)
	}
	if snap[len(snap)-1] != total {
		t.Fatalf("newest value = %d, want %d", snap[len(snap)-1], total)
	}
}

func TestBuffer_LatestSortedByTimestampDesc(t *testing.T) {
	t.Parallel()

	type timedValue struct {
		id int
		ts int64
	}

	b := New[timedValue](10, WithTimestamp(func(v timedValue) int64 { return v.ts }))
	b.Add(timedValue{id: 1, ts: 100})
	b.Add(timedValue{id: 2, ts: 300})
	b.Add(timedValue{id: 3, ts: 200})

	got := b.Latest(3)
	want := []int{2, 3, 1}
	if len(got) != len(want) {
		t.Fatalf("Latest(3) len = %d, want %d", len(got), len(want))
	}
	for i, value := range got {
		if value.id != want[i] {
			t.Fatalf("Latest(3)[%d] = %d, want %d", i, value.id, want[i])
		}
	}
}

func TestBuffer_LatestFewerThanRequested(t *testing.T) {
	t.Parallel()

	b := New[int](10, WithTimestamp(func(v int) int64 { return int64(v) }))
	b.Add(10)
	b.Add(20)

	got := b.Latest(5)
	if len(got) != 2 {
		t.Fatalf("Latest(5) len = %d, want 2", len(got))
	}
	if got[0] != 20 || got[1] != 10 {
		t.Fatalf("Latest(5) = %v, want [20 10]", got)
	}
}

func TestBuffer_LatestWithoutTimestampKeepsIngestionOrder(t *testing.T) {
	t.Parallel()

	b := New[int](10)
	b.Add(10)
	b.Add(20)
	b.Add(30)

	got := b.Latest(2)
	if len(got) != 2 {
		t.Fatalf("Latest(2) len = %d, want 2", len(got))
	}
	if got[0] != 20 || got[1] != 30 {
		t.Fatalf("Latest(2) = %v, want [20 30]", got)
	}
}

func TestBuffer_LatestEqualTimestampsKeepIngestionOrder(t *testing.T) {
	t.Parallel()

	b := New[int](10, WithTimestamp(func(int) int64 { return 1 }))
	b.Add(10)
	b.Add(20)
	b.Add(30)

	got := b.Latest(3)
	if len(got) != 3 {
		t.Fatalf("Latest(3) len = %d, want 3", len(got))
	}
	for i, want := range []int{10, 20, 30} {
		if got[i] != want {
			t.Fatalf("Latest(3)[%d] = %d, want %d", i, got[i], want)
		}
	}
}

func TestBuffer_LatestNonPositiveOrEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int
	}{
		{name: "zero", n: 0},
		{name: "negative", n: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := New[int](10, WithTimestamp(func(v int) int64 { return int64(v) }))
			if got := b.Latest(tt.n); len(got) != 0 {
				t.Fatalf("Latest(%d) len = %d, want 0", tt.n, len(got))
			}

			empty := New[int](10)
			if got := empty.Latest(5); len(got) != 0 {
				t.Fatalf("empty Latest(5) len = %d, want 0", len(got))
			}
		})
	}
}

func TestBuffer_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	b := New[int](100, WithTimestamp(func(v int) int64 { return int64(v) }))
	const writers = 5
	const perWriter = 200

	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		w := w
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWriter; i++ {
				b.Add(w*perWriter + i)
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
					if latest := b.Latest(10); len(latest) > 10 {
						t.Errorf("Latest(10) len out of range: %d", len(latest))
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

func TestTelemetryBuffer_AddAppendAndOverwrite(t *testing.T) {
	t.Parallel()

	b := NewTelemetryBuffer(3)
	for i := 1; i <= 5; i++ {
		b.Add([]byte(fmt.Sprintf(`{"device":"AE1","ts":%d,"battery":%d}`, i, i)))
	}

	if got := b.Len(); got != 3 {
		t.Fatalf("Len() = %d, want 3", got)
	}

	snap := b.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("Snapshot() len = %d, want 3", len(snap))
	}
	for i, wantTS := range []int64{3, 4, 5} {
		if got := payloadTS(t, snap[i]); got != wantTS {
			t.Fatalf("Snapshot()[%d] ts = %d, want %d", i, got, wantTS)
		}
	}
}

func TestTelemetryBuffer_LatestSortedByTSDesc(t *testing.T) {
	t.Parallel()

	b := NewTelemetryBuffer(10)
	b.Add([]byte(`{"device":"AE1","ts":100,"battery":1}`))
	b.Add([]byte(`{"device":"AE1","ts":300,"battery":3}`))
	b.Add([]byte(`{"device":"AE1","ts":200,"battery":2}`))

	got := b.Latest(3)
	if len(got) != 3 {
		t.Fatalf("Latest(3) len = %d, want 3", len(got))
	}

	for i, wantTS := range []int64{300, 200, 100} {
		var obj map[string]any
		if err := json.Unmarshal(got[i], &obj); err != nil {
			t.Fatalf("Latest(3)[%d] unmarshal: %v", i, err)
		}
		if gotTS := int64(obj["ts"].(float64)); gotTS != wantTS {
			t.Fatalf("Latest(3)[%d] ts = %d, want %d", i, gotTS, wantTS)
		}
		if _, ok := obj["device"]; !ok {
			t.Fatalf("Latest(3)[%d] missing original field device", i)
		}
	}
}

func TestTelemetryBuffer_LatestMissingTSSortsLast(t *testing.T) {
	t.Parallel()

	b := NewTelemetryBuffer(10)
	b.Add([]byte(`{"device":"AE1","battery":1}`))
	b.Add([]byte(`{"device":"AE1","ts":50}`))
	b.Add([]byte(`{"device":"AE1","ts":"bad"}`))

	got := b.Latest(3)
	if len(got) != 3 {
		t.Fatalf("Latest(3) len = %d, want 3", len(got))
	}
	if ts := payloadTS(t, got[0]); ts != 50 {
		t.Fatalf("Latest(3)[0] ts = %d, want 50", ts)
	}
	if string(got[0]) != `{"device":"AE1","ts":50}` {
		t.Fatalf("payload pass-through mismatch: %s", string(got[0]))
	}
	if string(got[1]) != `{"device":"AE1","battery":1}` {
		t.Fatalf("missing ts payload order mismatch: %s", string(got[1]))
	}
	if string(got[2]) != `{"device":"AE1","ts":"bad"}` {
		t.Fatalf("non-numeric ts payload order mismatch: %s", string(got[2]))
	}
}

func TestTelemetryBuffer_LatestReturnsCopies(t *testing.T) {
	t.Parallel()

	b := NewTelemetryBuffer(10)
	b.Add([]byte(`{"device":"AE1","ts":10}`))

	first := b.Latest(1)
	first[0][0] = 'X'

	second := b.Latest(1)
	if string(second[0]) != `{"device":"AE1","ts":10}` {
		t.Fatalf("stored payload mutated through Latest result: %s", string(second[0]))
	}
}

func TestTelemetryBuffer_NoPersistenceAcrossRestart(t *testing.T) {
	t.Parallel()

	prev := NewTelemetryBuffer(10)
	prev.Add([]byte(`{"device":"AE1","ts":1}`))
	if prev.Len() != 1 {
		t.Fatalf("pre-restart Len() = %d, want 1", prev.Len())
	}

	restarted := NewTelemetryBuffer(10)
	if restarted.Len() != 0 {
		t.Fatalf("post-restart Len() = %d, want 0", restarted.Len())
	}
	if got := restarted.Latest(5); len(got) != 0 {
		t.Fatalf("post-restart Latest(5) len = %d, want 0", len(got))
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := extractTS(json.RawMessage(tt.payload)); got != tt.want {
				t.Fatalf("extractTS(%s) = %d, want %d", tt.payload, got, tt.want)
			}
		})
	}
}

func payloadTS(t *testing.T, payload json.RawMessage) int64 {
	t.Helper()

	var obj struct {
		TS float64 `json:"ts"`
	}
	if err := json.Unmarshal(payload, &obj); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	return int64(obj.TS)
}
