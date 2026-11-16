package ingest

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"server/internal/buffer"
)

type scriptDialer struct {
	steps []dialStep
	i     atomic.Int32
	calls atomic.Int32
}

type dialStep struct {
	conn Conn
	err  error
}

func (d *scriptDialer) Dial(_ context.Context, _ string) (Conn, error) {
	idx := int(d.i.Add(1) - 1)
	d.calls.Add(1)
	if idx >= len(d.steps) {
		return nil, errors.New("no more dial steps")
	}
	step := d.steps[idx]
	if step.err != nil {
		return nil, step.err
	}
	return step.conn, nil
}

type scriptedConn struct {
	readErr error
	wait    time.Duration
	once    atomic.Bool
}

func (c *scriptedConn) Read(ctx context.Context) (Message, error) {
	if c.wait > 0 && c.once.CompareAndSwap(false, true) {
		t := time.NewTimer(c.wait)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return Message{}, ctx.Err()
		case <-t.C:
		}
	}
	if c.readErr != nil {
		return Message{}, c.readErr
	}
	<-ctx.Done()
	return Message{}, ctx.Err()
}

func (c *scriptedConn) Close(_ int, _ string) error { return nil }

func TestIngester_RunReconnectsWithBackoff(t *testing.T) {
	t.Parallel()

	buf := buffer.NewRingBuffer(10)
	// fail, fail, then connect and fail quickly to allow controlled retries.
	d := &scriptDialer{steps: []dialStep{
		{err: errors.New("dial 1")},
		{err: errors.New("dial 2")},
		{conn: &scriptedConn{readErr: errors.New("drop 1")}},
		{err: errors.New("dial 3")},
	}}
	base := NewSessionIngester(d, buf, slog.Default())
	ing := NewRetryingIngester(base, slog.Default())
	ing.baseDelay = 10 * time.Millisecond
	ing.maxDelay = 40 * time.Millisecond
	ing.successMin = time.Hour // treat short-lived sessions as not successful so backoff grows across retries
	ing.jitter = func(d time.Duration) time.Duration { return d }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var delays []time.Duration
	ing.sleep = func(ctx context.Context, d time.Duration) error {
		delays = append(delays, d)
		if len(delays) >= 3 {
			cancel()
			return ctx.Err()
		}
		return nil
	}

	err := ing.Run(ctx, "ws://example.test/telemetry")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	if len(delays) != 3 {
		t.Fatalf("sleep calls = %d, want 3 (%v)", len(delays), delays)
	}
	want := []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 40 * time.Millisecond}
	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delay[%d] = %v, want %v (all=%v)", i, delays[i], want[i], delays)
		}
	}
	if got := d.calls.Load(); got != 3 {
		t.Fatalf("dial calls = %d, want 3", got)
	}
}

func TestIngester_RunResetsBackoffAfterSuccessfulReconnect(t *testing.T) {
	t.Parallel()

	buf := buffer.NewRingBuffer(10)
	// 1) fail -> backoff 10ms (attempt becomes 1)
	// 2) connect and stay healthy then drop -> reset attempt to 0, backoff 10ms
	// 3) fail -> backoff 20ms (shows reset happened)
	d := &scriptDialer{steps: []dialStep{
		{err: errors.New("dial fail")},
		{conn: &scriptedConn{wait: 30 * time.Millisecond, readErr: errors.New("drop after healthy session")}},
		{err: errors.New("dial fail after healthy session")},
		{err: errors.New("dial fail again")},
	}}
	base := NewSessionIngester(d, buf, slog.Default())
	ing := NewRetryingIngester(base, slog.Default())
	ing.baseDelay = 10 * time.Millisecond
	ing.maxDelay = 80 * time.Millisecond
	ing.successMin = 20 * time.Millisecond
	ing.jitter = func(d time.Duration) time.Duration { return d }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var delays []time.Duration
	ing.sleep = func(ctx context.Context, d time.Duration) error {
		delays = append(delays, d)
		if len(delays) >= 3 {
			cancel()
			return ctx.Err()
		}
		return nil
	}

	err := ing.Run(ctx, "ws://example.test/telemetry")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run() error = %v, want context.Canceled", err)
	}
	want := []time.Duration{10 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond}
	if len(delays) != len(want) {
		t.Fatalf("sleep calls = %d, want %d (%v)", len(delays), len(want), delays)
	}
	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delay[%d] = %v, want %v (all=%v)", i, delays[i], want[i], delays)
		}
	}
}
