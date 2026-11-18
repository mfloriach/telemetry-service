package middleware

import (
	"net/http"
	"time"
)

// RequestTimeoutMiddleware enforces a per-request timeout.
func RequestTimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, `{"error":"request timeout"}`)
	}
}
