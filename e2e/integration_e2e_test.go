package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"server/internal/httpapi"
	"server/internal/ingest"
	"server/internal/utils/buffer"
	"server/internal/utils/kafka"
)

// mockKafkaConsumer is a test double for kafka.Consumer used in e2e tests.
type mockKafkaConsumer struct {
	mu     sync.Mutex
	msgs   [][]byte
	index  int
	err    error
	done   chan struct{}
}

func (c *mockKafkaConsumer) Consume(ctx context.Context) ([]byte, error) {
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

func (c *mockKafkaConsumer) Close() error {
	return nil
}

var _ kafka.Consumer = (*mockKafkaConsumer)(nil)

func TestEndToEnd_IngestReconnectAndQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	buf := buffer.NewRingBuffer(buffer.DefaultCapacity)
	consumer := &mockKafkaConsumer{
		msgs: [][]byte{
			[]byte(`{"satellite":"AE1","ts":100,"battery":1}`),
			[]byte(`{"satellite":"AE1","ts":200,"battery":2}`),
			[]byte(`{"satellite":"AE1","ts":150,"battery":1.5}`),
			[]byte(`{"satellite":"AE1","ts":300,"battery":3}`),
		},
	}
	sessionIng := ingest.NewSessionIngester(consumer, buf)
	svc := httpapi.NewService(buf)
	h := httpapi.NewHandler(svc, nil)
	apiSrv := httptest.NewServer(httpapi.NewMux(h))
	defer apiSrv.Close()

	ingestCtx, ingestCancel := context.WithCancel(ctx)
	defer ingestCancel()
	ingestErr := make(chan error, 1)
	go func() {
		ingestErr <- sessionIng.Run(ingestCtx)
	}()

	// Wait for all messages to be buffered from the Kafka consumer.
	waitForBuffered(ctx, t, buf, 4)

	resp, err := http.Get(apiSrv.URL + "/telemetry/latest?n=3")
	if err != nil {
		t.Fatalf("GET /telemetry/latest: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var got []struct {
		Satellite string  `json:"satellite"`
		TS        float64 `json:"ts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3: %#v", len(got), got)
	}
	wantTS := []float64{300, 200, 150}
	for i, want := range wantTS {
		if got[i].TS != want {
			t.Fatalf("record[%d].ts = %v, want %v (all=%#v)", i, got[i].TS, want, got)
		}
		if got[i].Satellite == "" {
			t.Fatalf("record[%d] missing pass-through field satellite: %#v", i, got[i])
		}
	}

	ingestCancel()
	select {
	case err := <-ingestErr:
		if err != nil && err != context.Canceled {
			t.Fatalf("ingester Run() error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for ingester shutdown")
	}
}

func waitForBuffered(ctx context.Context, t *testing.T, buf *buffer.RingBuffer, want int) {
	t.Helper()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if buf.Len() >= want {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatalf("timed out waiting for %d buffered records, got %d", want, buf.Len())
		case <-ticker.C:
		}
	}
}