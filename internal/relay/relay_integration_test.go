package relay

import (
	"context"
	"sync"
	"testing"

	"server/internal/utils/kafka"
	"server/internal/utils/ws"
)

// blockingConn returns messages one at a time, then blocks on Read until ctx is canceled.
type blockingConn struct {
	msgs []ws.Message
	i    int
	mu   sync.Mutex
}

func (c *blockingConn) Read(ctx context.Context) (ws.Message, error) {
	c.mu.Lock()
	if c.i >= len(c.msgs) {
		c.mu.Unlock()
		// Block until context is canceled
		<-ctx.Done()
		return ws.Message{}, ctx.Err()
	}
	m := c.msgs[c.i]
	c.i++
	c.mu.Unlock()
	return m, nil
}

func (c *blockingConn) Close(int, string) error { return nil }

// fakePublisher records published values for assertions.
type fakePublisher struct {
	mu        sync.Mutex
	published [][]byte
	done      chan struct{}
}

func (f *fakePublisher) Publish(ctx context.Context, value []byte) error {
	f.mu.Lock()
	f.published = append(f.published, value)
	count := len(f.published)
	f.mu.Unlock()
	if count == 3 {
		close(f.done)
	}
	return nil
}

func (f *fakePublisher) Close() error { return nil }

var _ kafka.Publisher = (*fakePublisher)(nil)

// failThenSucceedPublisher fails N times, then succeeds.
type failThenSucceedPublisher struct {
	mu        sync.Mutex
	failCount int
	published [][]byte
	attempts  int
	done      chan struct{}
}

func (f *failThenSucceedPublisher) Publish(ctx context.Context, value []byte) error {
	f.mu.Lock()
	f.attempts++
	if f.attempts <= f.failCount {
		f.mu.Unlock()
		return errPublishFailed
	}
	f.published = append(f.published, value)
	count := len(f.published)
	f.mu.Unlock()
	if count == 1 {
		close(f.done)
	}
	return nil
}

func (f *failThenSucceedPublisher) Close() error { return nil }

var errPublishFailed = errPublishFailedType{}

type errPublishFailedType struct{}

func (errPublishFailedType) Error() string { return "publish failed" }

// blockingDialer returns a blockingConn on each Dial.
type blockingDialer struct {
	msgs []ws.Message
}

func (d *blockingDialer) Dial(_ context.Context, _ string) (ws.Conn, error) {
	return &blockingConn{msgs: d.msgs}, nil
}

func TestRelay_Integration_WebSocketToKafkaMessageFlow(t *testing.T) {
	t.Parallel()

	dialer := &blockingDialer{
		msgs: []ws.Message{
			{Type: ws.MessageText, Data: []byte(`{"test": "message"}`)},
			{Type: ws.MessageText, Data: []byte(`{"temp": 23}`)},
			{Type: ws.MessageText, Data: []byte(`{"temp": 24}`)},
		},
	}

	publisher := &fakePublisher{done: make(chan struct{})}
	runner := NewRunner(dialer, publisher)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-publisher.done
		cancel()
	}()

	err := runner.Run(ctx, "ws://example.test/telemetry")
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}

	publisher.mu.Lock()
	defer publisher.mu.Unlock()

	if len(publisher.published) != 3 {
		t.Fatalf("expected 3 published messages, got %d", len(publisher.published))
	}

	expectedData := [][]byte{
		[]byte(`{"test": "message"}`),
		[]byte(`{"temp": 23}`),
		[]byte(`{"temp": 24}`),
	}
	for i, got := range publisher.published {
		if string(got) != string(expectedData[i]) {
			t.Fatalf("message %d: expected %q, got %q", i, expectedData[i], got)
		}
	}
}

func TestRelay_Integration_RetryOnPublishFailure(t *testing.T) {
	t.Parallel()

	dialer := &blockingDialer{
		msgs: []ws.Message{
			{Type: ws.MessageText, Data: []byte(`{"temp": 23}`)},
		},
	}

	failCount := 2
	publisher := &failThenSucceedPublisher{
		failCount: failCount,
		done:      make(chan struct{}),
	}
	runner := NewRunner(dialer, publisher)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-publisher.done
		cancel()
	}()

	err := runner.Run(ctx, "ws://example.test/telemetry")
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}

	publisher.mu.Lock()
	defer publisher.mu.Unlock()

	if len(publisher.published) != 1 {
		t.Fatalf("expected 1 published message after retries, got %d", len(publisher.published))
	}
	if string(publisher.published[0]) != `{"temp": 23}` {
		t.Fatalf("expected published data %q, got %q", `{"temp": 23}`, publisher.published[0])
	}

	if publisher.attempts <= 1 {
		t.Fatalf("expected at least %d publish attempts (failures + success), got %d", failCount+1, publisher.attempts)
	}
}

// Ensure blockingConn implements ws.Conn at compile time.
var _ ws.Conn = (*blockingConn)(nil)