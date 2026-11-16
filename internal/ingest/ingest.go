package ingest

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"
)

// Buffer stores ingested telemetry payloads.
type Buffer interface {
	Add(payload []byte)
}

// SessionRunner performs a single ingestion session.
type SessionRunner interface {
	RunOnce(ctx context.Context, url string) error
}

// SessionIngester connects to an upstream WebSocket and appends received messages to a buffer
// for a single session.
type SessionIngester struct {
	dialer Dialer
	buffer Buffer
	logger *slog.Logger
}

// NewSessionIngester creates a session ingester.
func NewSessionIngester(dialer Dialer, buffer Buffer, logger *slog.Logger) *SessionIngester {
	if logger == nil {
		logger = slog.Default()
	}
	return &SessionIngester{
		dialer: dialer,
		buffer: buffer,
		logger: logger,
	}
}

// RunOnce connects to url and reads messages until ctx is done or the connection fails.
// It returns ctx.Err() when canceled, otherwise the connection/read error.
func (i *SessionIngester) RunOnce(ctx context.Context, url string) error {
	conn, err := i.dialer.Dial(ctx, url)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}
	i.logger.Info("telemetry websocket connected", "url", url)
	defer func() {
		_ = conn.Close(1000, "closing")
	}()

	for {
		msg, err := conn.Read(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return ctx.Err()
			}
			return fmt.Errorf("read websocket: %w", err)
		}
		i.buffer.Add(msg.Data)
	}
}

// RetryingIngester maintains a resilient ingestion loop around a session runner,
// reconnecting with exponential backoff + jitter.
type RetryingIngester struct {
	runner     SessionRunner
	logger     *slog.Logger
	baseDelay  time.Duration
	maxDelay   time.Duration
	jitter     func(time.Duration) time.Duration
	sleep      func(context.Context, time.Duration) error
	successMin time.Duration
}

// NewRetryingIngester creates a retrying ingester decorator.
func NewRetryingIngester(runner SessionRunner, logger *slog.Logger) *RetryingIngester {
	if logger == nil {
		logger = slog.Default()
	}
	return &RetryingIngester{
		runner:     runner,
		logger:     logger,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   5 * time.Second,
		successMin: time.Second,
		jitter: func(d time.Duration) time.Duration {
			if d <= 0 {
				return 0
			}
			// Full jitter in [d/2, d] to avoid tight retry storms while keeping backoff bounded.
			half := d / 2
			if half <= 0 {
				return d
			}
			return half + rand.N(d-half+1)
		},
		sleep: sleepWithContext,
	}
}

// Run maintains a resilient ingestion loop, reconnecting with exponential backoff + jitter.
// It returns when ctx is canceled.
func (i *RetryingIngester) Run(ctx context.Context, url string) error {
	attempt := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		connectedAt := time.Now()
		err := i.runner.RunOnce(ctx, url)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ctx.Err()
		}

		if err != nil {
			i.logger.Warn("telemetry websocket disconnected", "url", url, "error", err)
			if time.Since(connectedAt) >= i.successMin {
				attempt = 0
			}
			delay := i.backoffDelay(attempt)
			attempt++
			i.logger.Info("telemetry websocket reconnect scheduled", "url", url, "attempt", attempt, "delay", delay.String())
			if sleepErr := i.sleep(ctx, delay); sleepErr != nil {
				return ctx.Err()
			}
			continue
		}
	}
}

func (i *RetryingIngester) backoffDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	exp := math.Pow(2, float64(attempt))
	d := time.Duration(float64(i.baseDelay) * exp)
	if d <= 0 || d > i.maxDelay {
		d = i.maxDelay
	}
	if i.jitter != nil {
		return i.jitter(d)
	}
	return d
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
