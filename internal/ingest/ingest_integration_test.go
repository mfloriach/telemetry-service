package ingest

import (
	"context"
	"sync"
	"testing"
	"time"

	"server/internal/utils/kafka"
)

// recordingBuffer is a Buffer implementation that records all added payloads.
type recordingBuffer struct {
	mu      sync.Mutex
	payloads [][]byte
}

func (b *recordingBuffer) Add(payload []byte) {
	b.mu.Lock()
	b.payloads = append(b.payloads, payload)
	b.mu.Unlock()
}

func (b *recordingBuffer) getPayloads() [][]byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make([][]byte, len(b.payloads))
	copy(result, b.payloads)
	return result
}

// mockConsumer is a test double for kafka.Consumer that returns messages one at a time,
// then blocks until ctx is canceled.
type mockConsumer struct {
	mu      sync.Mutex
	msgs    [][]byte
	index   int
	err     error
	closeErr error
	done    chan struct{}
}

func (c *mockConsumer) Consume(ctx context.Context) ([]byte, error) {
	c.mu.Lock()
	if c.index >= len(c.msgs) {
		c.mu.Unlock()
		// Block until context is canceled
		<-ctx.Done()
		return nil, ctx.Err()
	}
	msg := c.msgs[c.index]
	c.index++
	c.mu.Unlock()
	return msg, nil
}

func (c *mockConsumer) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.closeErr
}

var _ kafka.Consumer = (*mockConsumer)(nil)

func TestIngest_Integration_BufferRetainsMessagesUnderUnavailability(t *testing.T) {
	t.Parallel()

	buffer := &recordingBuffer{}
	consumer := &mockConsumer{
		msgs: [][]byte{
			[]byte(`{"temp": 23}`),
			[]byte(`{"temp": 24}`),
			[]byte(`{"temp": 25}`),
		},
	}
	ingester := NewSessionIngester(consumer, buffer)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		for len(buffer.getPayloads()) < 3 {
			// wait for all messages to be buffered
		}
		close(done)
	}()

	go func() {
		<-done
		cancel()
	}()

	err := ingester.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}

	payloads := buffer.getPayloads()
	if len(payloads) != 3 {
		t.Fatalf("expected 3 buffered payloads, got %d", len(payloads))
	}

	expectedData := [][]byte{
		[]byte(`{"temp": 23}`),
		[]byte(`{"temp": 24}`),
		[]byte(`{"temp": 25}`),
	}
	for i, got := range payloads {
		if string(got) != string(expectedData[i]) {
			t.Fatalf("payload %d: expected %q, got %q", i, expectedData[i], got)
		}
	}
}

func TestIngest_Integration_ContextCancellationDuringRead(t *testing.T) {
	t.Parallel()

	buffer := &recordingBuffer{}
	consumer := &mockConsumer{
		msgs: [][]byte{
			[]byte(`{"temp": 23}`),
		},
	}
	ingester := NewSessionIngester(consumer, buffer)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := ingester.Run(ctx)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}

	// Verify at least the first message was buffered before cancellation
	payloads := buffer.getPayloads()
	if len(payloads) < 1 {
		t.Fatal("expected at least 1 buffered payload before cancellation")
	}
}