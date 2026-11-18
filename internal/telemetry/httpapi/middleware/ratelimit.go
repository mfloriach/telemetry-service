package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

// RateLimiter provides a simple in-memory fixed-window rate limiter.
type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]rateWindow
}

type rateWindow struct {
	count int
	reset time.Time
}

// NewRateLimiter creates a fixed-window rate limiter.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 60
	}
	if window <= 0 {
		window = time.Minute
	}
	return &RateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]rateWindow),
	}
}

// Allow reports whether a request from key is allowed.
func (rl *RateLimiter) Allow(key string) bool {
	now := time.Now()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, ok := rl.clients[key]
	if !ok || now.After(entry.reset) {
		rl.clients[key] = rateWindow{count: 1, reset: now.Add(rl.window)}
		return true
	}
	if entry.count >= rl.limit {
		return false
	}
	entry.count++
	rl.clients[key] = entry
	return true
}

// RateLimitMiddleware limits requests per client key (remote IP).
func RateLimitMiddleware(limiter *RateLimiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter != nil && !limiter.Allow(clientKey(r)) {
				writeError(w, r, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientKey(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
