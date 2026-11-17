package kafka

import (
	"context"
	"testing"
)

// mockConsumer is a test double for the Consumer interface.
type mockConsumer struct {
	messages [][]byte
	err      error
	closeErr error
	index    int
}

func (m *mockConsumer) Consume(_ context.Context) ([]byte, error) {
	if m.index >= len(m.messages) {
		return nil, m.err
	}
	msg := m.messages[m.index]
	m.index++
	return msg, nil
}

func (m *mockConsumer) Close() error {
	return m.closeErr
}

// Ensure mockConsumer satisfies Consumer at compile time.
var _ Consumer = (*mockConsumer)(nil)

func TestMockConsumer_SatisfiesConsumerInterface(t *testing.T) {
	t.Parallel()

	var _ Consumer = &mockConsumer{}

	m := &mockConsumer{
		messages: [][]byte{[]byte("msg1"), []byte("msg2")},
	}

	msg1, err := m.Consume(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(msg1) != "msg1" {
		t.Fatalf("expected msg1, got %q", msg1)
	}

	msg2, err := m.Consume(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(msg2) != "msg2" {
		t.Fatalf("expected msg2, got %q", msg2)
	}

	_, err = m.Consume(context.Background())
	if err != nil {
		t.Fatalf("expected no error after messages exhausted, got %v", err)
	}

	if err := m.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
}