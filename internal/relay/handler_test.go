package relay

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

type recordedPublish struct {
	value []byte
}

type fakePublisher struct {
	mu      sync.Mutex
	records []recordedPublish
	err     error
}

func (f *fakePublisher) Publish(_ context.Context, value []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.records = append(f.records, recordedPublish{value: append([]byte(nil), value...)})
	return f.err
}

func (f *fakePublisher) published() []recordedPublish {
	f.mu.Lock()
	defer f.mu.Unlock()

	return append([]recordedPublish(nil), f.records...)
}

func TestNewHandler_RequiresPublisher(t *testing.T) {
	t.Parallel()

	handler, err := NewHandler(nil)
	if err == nil || !strings.Contains(err.Error(), "publisher") {
		t.Fatalf("NewHandler(nil) error = %v, want containing %q", err, "publisher")
	}
	if handler != nil {
		t.Fatal("NewHandler(nil) returned a handler")
	}
}

func TestHandler_Handle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		messageType string
		payload     string
		publishErr  error
		wantErr     string
	}{
		{
			name:        "text payload is published unchanged",
			messageType: "text",
			payload:     `{"device_id":"d-1","value":42}`,
		},
		{
			name:        "binary payload is published unchanged",
			messageType: "binary",
			payload:     "\x00\x01\x02",
		},
		{
			name:        "empty payload is published unchanged",
			messageType: "text",
			payload:     "",
		},
		{
			name:        "publish failure is returned with context",
			messageType: "text",
			payload:     `{"device_id":"d-2"}`,
			publishErr:  errors.New("kafka unavailable"),
			wantErr:     "publish telemetry record",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			publisher := &fakePublisher{err: tt.publishErr}
			handler, err := NewHandler(publisher)
			if err != nil {
				t.Fatalf("NewHandler() unexpected error: %v", err)
			}

			err = handler.Handle(context.Background(), tt.messageType, []byte(tt.payload))
			records := publisher.published()

			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Handle() error = %v, want containing %q", err, tt.wantErr)
				}
				if !errors.Is(err, tt.publishErr) {
					t.Fatalf("Handle() error = %v, want wrapping %v", err, tt.publishErr)
				}
				if len(records) != 1 || string(records[0].value) != tt.payload {
					t.Fatalf("published records = %#v, want one record with the original payload", records)
				}
				return
			}

			if err != nil {
				t.Fatalf("Handle() unexpected error: %v", err)
			}
			if len(records) != 1 {
				t.Fatalf("published records = %d, want 1", len(records))
			}
			if got := string(records[0].value); got != tt.payload {
				t.Fatalf("published payload = %q, want %q", got, tt.payload)
			}
		})
	}
}
