package relay

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"server/internal/ingest"
)

type fakeConn struct {
	msgs []ingest.Message
	i    int
}

func (c *fakeConn) Read(_ context.Context) (ingest.Message, error) {
	if c.i >= len(c.msgs) {
		return ingest.Message{}, errors.New("stream closed")
	}
	m := c.msgs[c.i]
	c.i++
	return m, nil
}

func (c *fakeConn) Close(int, string) error { return nil }

type fakeDialer struct {
	conn ingest.Conn
	err  error
}

func (d fakeDialer) Dial(context.Context, string) (ingest.Conn, error) {
	if d.err != nil {
		return nil, d.err
	}
	return d.conn, nil
}

type fakePublisher struct {
	values [][]byte
	fail   error
}

func (p *fakePublisher) Publish(_ context.Context, value []byte) error {
	if p.fail != nil {
		return p.fail
	}
	cp := make([]byte, len(value))
	copy(cp, value)
	p.values = append(p.values, cp)
	return nil
}

func (p *fakePublisher) Close() error { return nil }

func TestRunner_PublishesReceivedRecords(t *testing.T) {
	t.Parallel()

	pub := &fakePublisher{}
	r := NewRunner(fakeDialer{conn: &fakeConn{msgs: []ingest.Message{{Type: ingest.MessageText, Data: []byte(`{"ts":1}`)}}}}, pub, slog.New(slog.NewTextHandler(io.Discard, nil)))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_ = r.Run(ctx, "ws://example.test")

	if len(pub.values) != 1 || string(pub.values[0]) != `{"ts":1}` {
		t.Fatalf("expected published original record, got %#v", pub.values)
	}
}

func TestRunner_ReconnectsAfterDialFailure(t *testing.T) {
	t.Parallel()

	var dials atomic.Int32
	pub := &fakePublisher{}
	r := NewRunner(dialerFunc(func(context.Context, string) (ingest.Conn, error) {
		dials.Add(1)
		return nil, errors.New("dial failed")
	}), pub, slog.New(slog.NewTextHandler(io.Discard, nil)))
	r.baseDelay = time.Millisecond
	r.maxDelay = 2 * time.Millisecond
	r.jitter = func(d time.Duration) time.Duration { return d }

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_ = r.Run(ctx, "ws://example.test")

	if dials.Load() < 2 {
		t.Fatalf("expected reconnect attempts, got %d dials", dials.Load())
	}
}

type dialerFunc func(context.Context, string) (ingest.Conn, error)

func (f dialerFunc) Dial(ctx context.Context, url string) (ingest.Conn, error) { return f(ctx, url) }
