package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httprate"
)

// Per-IP request budgets. The general limit is loose enough that an interactive
// SPA session never notices it; the auth limit is tight because each
// /api/auth/callback fans out to token and JWKS requests against the IdP.
const (
	APIRateLimitRequests = 100
	APIRateLimitWindow   = time.Minute

	AuthRateLimitRequests = 10
	AuthRateLimitWindow   = time.Minute
)

// clientIPKey keys a rate limiter on the client IP as chi's RealIP middleware
// left it in r.RemoteAddr (resolved from X-Forwarded-For / X-Real-IP). The
// limit is therefore only as sound as the reverse proxy setting those headers;
// on a directly-exposed deployment it keys on the real socket address, which is
// also fine. CanonicalizeIP collapses an IPv6 client to its /64 so it can't
// rotate addresses to win a fresh bucket per request.
func clientIPKey(r *http.Request) (string, error) {
	ip := r.RemoteAddr
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}
	return httprate.CanonicalizeIP(ip), nil
}

func limitJSON429(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	_, _ = w.Write([]byte(`{"message":"rate limit exceeded"}`))
}

// RateLimitByIP limits requests per client IP over the given window and answers
// a breach with a JSON 429 shaped like the rest of the API.
func RateLimitByIP(requests int, window time.Duration) func(http.Handler) http.Handler {
	return httprate.LimitBy(requests, window, clientIPKey, httprate.WithLimitHandler(limitJSON429))
}

// RateLimitByIPForPrefixes applies a per-IP rate limit only to requests whose
// path starts with one of prefixes; every other request passes straight
// through. Used to put a tighter budget on the auth routes without touching the
// rest of the API.
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
