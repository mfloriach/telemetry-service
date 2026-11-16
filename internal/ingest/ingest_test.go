package ingest

import (
	"context"
	"errors"
	"testing"
	"time"

	"server/internal/utils/buffer"
)

type fakeDialer struct {
	conn Conn
	err  error
	url  string
}

func (f *fakeDialer) Dial(_ context.Context, url string) (Conn, error) {
	f.url = url
	if f.err != nil {
		return nil, f.err
	}
	return f.conn, nil
}

type fakeConn struct {
	msgs    []Message
	i       int
	readErr error
	closed  bool
}

func (c *fakeConn) Read(ctx context.Context) (Message, error) {
	if c.i < len(c.msgs) {
		m := c.msgs[c.i]
		c.i++
		return m, nil
	}
	if c.readErr != nil {
		return Message{}, c.readErr
	}
	<-ctx.Done()
	return Message{}, ctx.Err()
}

func (c *fakeConn) Close(_ int, _ string) error {
	c.closed = true
	return nil
}

func TestIngester_RunOnceBuffersReceivedRecords(t *testing.T) {
	t.Parallel()

	buf := buffer.NewRingBuffer(10)
	conn := &fakeConn{
		msgs: []Message{
			{Type: MessageText, Data: []byte(`{"satellite":"AE1","ts":1735657200000,"battery":11.1}`)},
			{Type: MessageText, Data: []byte(`{"satellite":"AE1","ts":1735657200100,"battery":22.2}`)},
		},
		readErr: errors.New("disconnect"),
	}
	dialer := &fakeDialer{conn: conn}
	ing := NewSessionIngester(dialer, buf, nil)

	err := ing.RunOnce(context.Background(), "ws://example.test/telemetry")
	if err == nil {
		t.Fatal("RunOnce() error = nil, want disconnect error")
	}
	if dialer.url != "ws://example.test/telemetry" {
		t.Fatalf("dial url = %q, want ws://example.test/telemetry", dialer.url)
	}
	if !conn.closed {
		t.Fatal("expected connection to be closed")
	}

	latest := buf.Latest(2)
	if len(latest) != 2 {
		t.Fatalf("buffered records = %d, want 2", len(latest))
	}
	if got := string(latest[0].Payload); got != `{"satellite":"AE1","ts":1735657200100,"battery":22.2}` {
		t.Fatalf("latest[0] payload = %s, want pass-through JSON", got)
	}
	if got := string(latest[1].Payload); got != `{"satellite":"AE1","ts":1735657200000,"battery":11.1}` {
		t.Fatalf("latest[1] payload = %s, want pass-through JSON", got)
	}
}

func TestIngester_RunOnceDialError(t *testing.T) {
	t.Parallel()

	buf := buffer.NewRingBuffer(10)
	dialer := &fakeDialer{err: errors.New("dial failed")}
	ing := NewSessionIngester(dialer, buf, nil)

	err := ing.RunOnce(context.Background(), "ws://example.test/telemetry")
	if err == nil {
		t.Fatal("RunOnce() error = nil, want dial error")
	}
	if buf.Len() != 0 {
		t.Fatalf("buffer len = %d, want 0", buf.Len())
	}
}

func TestIngester_RunOnceContextCanceled(t *testing.T) {
	t.Parallel()

	buf := buffer.NewRingBuffer(10)
	conn := &fakeConn{}
	dialer := &fakeDialer{conn: conn}
	ing := NewSessionIngester(dialer, buf, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := ing.RunOnce(ctx, "ws://example.test/telemetry")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunOnce() error = %v, want context.Canceled", err)
	}

	// also verify blocking read path respects cancellation
	ctx2, cancel2 := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel2()
	conn2 := &fakeConn{}
	ing2 := NewSessionIngester(&fakeDialer{conn: conn2}, buf, nil)
	err = ing2.RunOnce(ctx2, "ws://example.test/telemetry")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunOnce() timeout error = %v, want context.DeadlineExceeded", err)
	}
}
