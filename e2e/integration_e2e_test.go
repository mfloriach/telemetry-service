package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"server/internal/httpapi"
	"server/internal/ingest"
	"server/internal/utils/buffer"
)

func TestEndToEnd_IngestReconnectAndQuery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Upstream WebSocket server that accepts one connection at a time and can drop clients.
	acceptCh := make(chan *websocket.Conn, 4)
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		acceptCh <- c
		// Keep handler alive until test closes connection/server.
		<-r.Context().Done()
	}))
	defer upstream.Close()
	wsURL := "ws" + strings.TrimPrefix(upstream.URL, "http")

	buf := buffer.NewRingBuffer(buffer.DefaultCapacity)
	sessionIng := ingest.NewSessionIngester(ingest.CoderDialer{}, buf, nil)
	ing := ingest.NewRetryingIngester(sessionIng, nil)
	svc := httpapi.NewService(buf)
	h := httpapi.NewHandler(svc, nil)
	apiSrv := httptest.NewServer(httpapi.NewMux(h))
	defer apiSrv.Close()

	ingestCtx, ingestCancel := context.WithCancel(ctx)
	defer ingestCancel()
	ingestErr := make(chan error, 1)
	go func() {
		ingestErr <- ing.Run(ingestCtx, wsURL)
	}()

	// First connection: send some records, then simulate disconnect.
	var conn1 *websocket.Conn
	select {
	case conn1 = <-acceptCh:
	case <-ctx.Done():
		t.Fatal("timed out waiting for first upstream connection")
	}
	writeJSON(ctx, t, conn1, `{"satellite":"AE1","ts":100,"battery":1}`)
	writeJSON(ctx, t, conn1, `{"satellite":"AE1","ts":200,"battery":2}`)
	waitForBuffered(ctx, t, buf, 2)
	_ = conn1.Close(websocket.StatusInternalError, "simulated drop")

	// Reconnect: ingester should establish a new connection.
	var conn2 *websocket.Conn
	select {
	case conn2 = <-acceptCh:
	case <-ctx.Done():
		t.Fatal("timed out waiting for reconnect to upstream")
	}
	writeJSON(ctx, t, conn2, `{"satellite":"AE1","ts":150,"battery":1.5}`)
	writeJSON(ctx, t, conn2, `{"satellite":"AE1","ts":300,"battery":3}`)
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
	_ = conn2.Close(websocket.StatusNormalClosure, "test done")
}

func writeJSON(ctx context.Context, t *testing.T, c *websocket.Conn, payload string) {
	t.Helper()
	if err := c.Write(ctx, websocket.MessageText, []byte(payload)); err != nil {
		t.Fatalf("write upstream message: %v", err)
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
