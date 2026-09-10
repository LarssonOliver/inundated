package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httprate"
)

const (
	APIRateLimitRequests = 100
	APIRateLimitWindow   = time.Minute

	AuthRateLimitRequests = 10
	AuthRateLimitWindow   = time.Minute
)

func clientIPKey(r *http.Request) (string, error) {
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	return httprate.CanonicalizeIP(ip), nil
}

func limitJSON429(w http.ResponseWriter, r *http.Request) {
	key, _ := clientIPKey(r)
	slog.WarnContext(r.Context(), "rate limit exceeded; dropping request",
		slog.String("client_ip", key),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(`{"message":"rate limit exceeded"}`))
}

func RateLimitByIP(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(requests, window, clientIPKey, httprate.WithLimitHandler(limitJSON429))
}

func RateLimitByIPForPrefixes(requests int, window time.Duration, prefixes ...string) func(http.Handler) http.Handler {
	limited := RateLimitByIP(requests, window)
	return func(next http.Handler) http.Handler {
		guarded := limited(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range prefixes {
				if strings.HasPrefix(r.URL.Path, p) {
					guarded.ServeHTTP(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
