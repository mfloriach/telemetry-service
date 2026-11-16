package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"server/internal/buffer"
)

func TestService_Latest(t *testing.T) {
	t.Parallel()

	t.Run("returns payloads sorted desc by ts", func(t *testing.T) {
		t.Parallel()
		buf := buffer.NewRingBuffer(10)
		buf.Add([]byte(`{"satellite":"AE1","ts":10}`))
		buf.Add([]byte(`{"satellite":"AE1","ts":30}`))
		buf.Add([]byte(`{"satellite":"AE1","ts":20}`))

		svc := NewService(buf)
		got, err := svc.Latest(context.Background(), 2)
		if err != nil {
			t.Fatalf("Latest() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len = %d, want 2", len(got))
		}
		var first, second struct {
			TS float64 `json:"ts"`
		}
		if err := json.Unmarshal(got[0], &first); err != nil {
			t.Fatalf("unmarshal first: %v", err)
		}
		if err := json.Unmarshal(got[1], &second); err != nil {
			t.Fatalf("unmarshal second: %v", err)
		}
		if int64(first.TS) != 30 || int64(second.TS) != 20 {
			t.Fatalf("unexpected order: %v %v", first.TS, second.TS)
		}
	})

	t.Run("invalid n returns ErrInvalidLimit", func(t *testing.T) {
		t.Parallel()
		svc := NewService(buffer.NewRingBuffer(10))
		for _, n := range []int{0, -1, 101} {
			_, err := svc.Latest(context.Background(), n)
			if !errors.Is(err, ErrInvalidLimit) {
				t.Fatalf("Latest(%d) error = %v, want ErrInvalidLimit", n, err)
			}
		}
	})
}
