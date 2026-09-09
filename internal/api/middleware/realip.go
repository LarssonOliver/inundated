package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// RealIP rewrites r.RemoteAddr to the client address reported by a trusted
// reverse proxy.
//
// The forwarded headers are consulted only when the direct TCP peer
// (r.RemoteAddr) falls inside one of trustedProxies. Otherwise -- including
// when trustedProxies is empty -- r.RemoteAddr is left untouched, so anything
// keying off it (the per-IP rate limiter) sees the real connection and cannot
// be spoofed by a client sending its own X-Forwarded-For.
//
// When the peer is trusted, X-Forwarded-For is walked right-to-left and the
// first address that is not itself a trusted proxy is taken as the client; if
// every entry is trusted, the left-most (original) entry is used. X-Real-IP is
// the fallback when X-Forwarded-For is absent.
func RealIP(trustedProxies []netip.Prefix) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip := realIP(r, trustedProxies); ip != "" {
				r.RemoteAddr = ip
			}
			next.ServeHTTP(w, r)
		})
	}
}

func realIP(r *http.Request, trusted []netip.Prefix) string {
	if len(trusted) == 0 {
		return ""
	}

	peer := parseAddr(r.RemoteAddr)
	if !peer.IsValid() || !isTrusted(peer, trusted) {
		return ""
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		var leftmost netip.Addr
		for i := len(parts) - 1; i >= 0; i-- {
			addr := parseAddr(strings.TrimSpace(parts[i]))
			if !addr.IsValid() {
				continue
			}
			leftmost = addr
			if !isTrusted(addr, trusted) {
				return addr.String()
			}
		}
		if leftmost.IsValid() {
			return leftmost.String()
		}
	}

	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		if addr := parseAddr(xrip); addr.IsValid() {
			return addr.String()
		}
	}

	return ""
}

func parseAddr(s string) netip.Addr {
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}
	}
	return addr.Unmap()
}

func isTrusted(addr netip.Addr, trusted []netip.Prefix) bool {
	for _, p := range trusted {
		if p.Contains(addr) {
			return true
		}
	}
	return false
}
