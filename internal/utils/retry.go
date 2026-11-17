package utils

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"
)

// SessionRunner performs a single ingestion session.
type SessionRunner interface {
	RunOnce(ctx context.Context, url string) error
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
			return half + rand.N(d-half+1) //nolint:gosec // non-cryptographic jitter for reconnect backoff
		},
		sleep: sleepWithContext,
	}
}

// SetBaseDelay sets the base delay for the backoff strategy.
func (i *RetryingIngester) SetBaseDelay(d time.Duration) {
	i.baseDelay = d
}

// SetMaxDelay sets the maximum delay for the backoff strategy.
func (i *RetryingIngester) SetMaxDelay(d time.Duration) {
	i.maxDelay = d
}

// SetSuccessMin sets the minimum success duration before resetting attempt counter.
func (i *RetryingIngester) SetSuccessMin(d time.Duration) {
	i.successMin = d
}

// SetJitter sets the jitter function for the backoff strategy.
func (i *RetryingIngester) SetJitter(j func(d time.Duration) time.Duration) {
	i.jitter = j
}

// SetSleep sets the sleep function for the backoff strategy.
func (i *RetryingIngester) SetSleep(s func(context.Context, time.Duration) error) {
	i.sleep = s
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
			i.logger.WarnContext(ctx, "telemetry websocket disconnected", "url", url, "error", err)
			if time.Since(connectedAt) >= i.successMin {
				attempt = 0
			}
			delay := i.backoffDelay(attempt)
			attempt++
			i.logger.InfoContext(ctx, "telemetry websocket reconnect scheduled", "url", url, "attempt", attempt, "delay", delay.String())
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

// SleepWithContext sleeps for the given duration, respecting context cancellation.
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

// backoffDelay calculates exponential backoff with the given attempt number, base delay, and max delay.
func backoffDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	exp := math.Pow(2, float64(attempt))
	d := time.Duration(float64(baseDelay) * exp)
	if d <= 0 || d > maxDelay {
		d = maxDelay
	}
	return d
}

// WithRetry applies exponential backoff + jitter to the given function.
// It retries until ctx is canceled, the function succeeds, or the error is not retriable.
func WithRetry(ctx context.Context, fn func(ctx context.Context) error, baseDelay, maxDelay time.Duration, jitter func(d time.Duration) time.Duration, successMin time.Duration) error {
	attempt := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn(ctx)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ctx.Err()
		}

		if err != nil {
			if time.Since(time.Now()) >= successMin {
				attempt = 0
			}
			delay := backoffDelay(attempt, baseDelay, maxDelay)
			if jitter != nil {
				delay = jitter(delay)
			}
			if sleepErr := sleepWithContext(ctx, delay); sleepErr != nil {
				return ctx.Err()
			}
			attempt++
			continue
		}
		return nil
	}
}