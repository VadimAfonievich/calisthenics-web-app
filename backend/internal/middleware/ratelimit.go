package middleware

import (
	"net"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimit applies a fixed-window limit per client IP. Redis errors fail open so
// a transient cache outage does not take the API down.
func RateLimit(client redis.UniversalClient, limit int, window time.Duration, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
			ip = host
		}
		key := "rate:" + time.Now().UTC().Format("200601021504") + ":" + ip
		count, err := client.Incr(r.Context(), key).Result()
		if err == nil && count == 1 {
			_ = client.Expire(r.Context(), key, window).Err()
		}
		if err == nil && count > int64(limit) {
			w.Header().Set("Retry-After", "60")
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"code":"RATE_LIMITED","message":"Too many requests; try again shortly"}}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func LimitBody(maxBytes int64, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
		}
		next.ServeHTTP(w, r)
	})
}
