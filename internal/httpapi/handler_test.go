package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"server/internal/buffer"
	"server/internal/config"
	"server/internal/httpapi/middleware"
	"server/internal/logging"
)

func TestHandleLatest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		seed       []string
		wantStatus int
		wantLen    int
		wantTSDesc []int64
		wantErr    string
	}{
		{
			name:       "valid request returns latest n sorted desc by ts",
			query:      "?n=3",
			seed:       []string{`{"satellite":"AE1","ts":10,"battery":1}`, `{"satellite":"AE1","ts":30,"battery":3}`, `{"satellite":"AE1","ts":20,"battery":2}`, `{"satellite":"AE1","ts":40,"battery":4}`},
			wantStatus: http.StatusOK,
			wantLen:    3,
			wantTSDesc: []int64{40, 30, 20},
		},
		{
			name:       "fewer than n returns all available",
			query:      "?n=5",
			seed:       []string{`{"satellite":"AE1","ts":10}`, `{"satellite":"AE1","ts":20}`},
			wantStatus: http.StatusOK,
			wantLen:    2,
			wantTSDesc: []int64{20, 10},
		},
		{name: "missing n", query: "", wantStatus: http.StatusBadRequest, wantErr: "invalid query parameter: n must be an integer between 1 and 100"},
		{name: "non integer n", query: "?n=abc", wantStatus: http.StatusBadRequest, wantErr: "invalid query parameter: n must be an integer between 1 and 100"},
		{name: "zero n", query: "?n=0", wantStatus: http.StatusBadRequest, wantErr: "invalid query parameter: n must be an integer between 1 and 100"},
		{name: "negative n", query: "?n=-1", wantStatus: http.StatusBadRequest, wantErr: "invalid query parameter: n must be an integer between 1 and 100"},
		{name: "too large n", query: "?n=101", wantStatus: http.StatusBadRequest, wantErr: "invalid query parameter: n must be an integer between 1 and 100"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := buffer.NewRingBuffer(100)
			for _, s := range tt.seed {
				buf.Add([]byte(s))
			}
			h := NewHandler(NewService(buf), nil)
			mux := http.NewServeMux()
			h.RegisterRoutes(mux)

			req := httptest.NewRequest(http.MethodGet, "/telemetry/latest"+tt.query, nil)
			rr := httptest.NewRecorder()
			mux.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body=%q", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", ct)
			}
			if tt.wantStatus != http.StatusOK {
				var errBody ErrorResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &errBody); err != nil {
					t.Fatalf("unmarshal error response: %v body=%q", err, rr.Body.String())
				}
				if errBody.Error != tt.wantErr {
					t.Fatalf("error = %q, want %q", errBody.Error, tt.wantErr)
				}
				return
			}

			var got []struct {
				Satellite string  `json:"satellite"`
				TS        float64 `json:"ts"`
			}
			if err := json.Unmarshal(rr.Body.Bytes(), &got); err != nil {
				t.Fatalf("unmarshal response: %v body=%q", err, rr.Body.String())
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tt.wantLen)
			}
			for i, wantTS := range tt.wantTSDesc {
				if int64(got[i].TS) != wantTS {
					t.Fatalf("record[%d] ts = %v, want %d", i, got[i].TS, wantTS)
				}
				if got[i].Satellite == "" {
					t.Fatalf("record[%d] missing pass-through field satellite: %#v", i, got[i])
				}
			}
		})
	}
}

func TestNewServer_EndpointReachable(t *testing.T) {
	t.Parallel()

	buf := buffer.NewRingBuffer(10)
	buf.Add([]byte(`{"satellite":"AE1","ts":20}`))
	buf.Add([]byte(`{"satellite":"AE1","ts":10}`))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	srv := NewServer(addr, NewHandler(NewService(buf), nil), nil)
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	var resp *http.Response
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		req, reqErr := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://"+addr+"/telemetry/latest?n=1", nil)
		if reqErr != nil {
			t.Fatalf("new request: %v", reqErr)
		}
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		_ = srv.Shutdown(context.Background())
		t.Fatalf("GET /telemetry/latest: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if err := <-errCh; err != nil && err != http.ErrServerClosed {
		t.Fatalf("ListenAndServe: %v", err)
	}
}

func TestNewServer_RequestLogsIncludeRequestID(t *testing.T) {
	t.Parallel()

	buf := buffer.NewRingBuffer(10)
	buf.Add([]byte(`{"satellite":"AE1","ts":20}`))

	var logBuf bytes.Buffer
	logger := logging.NewLogger(config.Config{LogLevel: "info", LogFormat: "json"}, &logBuf, middleware.RequestIDFromContext)
	srv := NewServer(":0", NewHandler(NewService(buf), logger), logger)

	req := httptest.NewRequest(http.MethodGet, "/telemetry/latest?n=1", nil)
	req.Header.Set("X-Request-Id", "req-test-123")
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%q", rr.Code, http.StatusOK, rr.Body.String())
	}

	found := false
	for _, line := range strings.Split(strings.TrimSpace(logBuf.String()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("unmarshal log line %q: %v", line, err)
		}
		msg, _ := rec["msg"].(string)
		if msg != "http request" {
			continue
		}
		if got, _ := rec["request_id"].(string); got != "req-test-123" {
			t.Fatalf("request_id = %q, want %q; record=%v", got, "req-test-123", rec)
		}
		found = true
	}
	if !found {
		t.Fatalf("expected http request log with request_id; logs=%q", logBuf.String())
	}
}
