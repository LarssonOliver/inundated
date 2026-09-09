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
// be spoofed by a client sending its own forwarded-for header.
//
// forwardHeaders names the headers to check, in priority order; the first that
// yields a valid address wins. "X-Forwarded-For" is chain-aware: all header
// fields are joined, walked right-to-left, and the first hop that is not itself
// a trusted proxy is taken as the client (falling back to the left-most entry
// if every hop is trusted). Any other configured header is a single value taken
// as-is. Callers pass a validated list; unrecognised names are treated as
// single-value headers.
func RealIP(trustedProxies []netip.Prefix, forwardHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ip := realIP(r, trustedProxies, forwardHeaders); ip != "" {
				r.RemoteAddr = ip
			}
			next.ServeHTTP(w, r)
		})
	}
}

func realIP(r *http.Request, trusted []netip.Prefix, headers []string) string {
	if len(trusted) == 0 {
		return ""
	}

	peer := parseAddr(r.RemoteAddr)
	if !peer.IsValid() || !isTrusted(peer, trusted) {
		return ""
	}

	if len(headers) == 0 {
		headers = []string{"X-Forwarded-For"}
	}

	for _, h := range headers {
		if strings.EqualFold(h, "X-Forwarded-For") {
			if ip := fromForwardedFor(r, trusted); ip != "" {
				return ip
			}
			continue
		}
		if v := strings.TrimSpace(r.Header.Get(h)); v != "" {
			if addr := parseAddr(v); addr.IsValid() {
				return addr.String()
			}
		}
	}

	return ""
}

// fromForwardedFor resolves the client address from every X-Forwarded-For
// header field, joined into one chain and walked right-to-left.
func fromForwardedFor(r *http.Request, trusted []netip.Prefix) string {
	fields := r.Header.Values("X-Forwarded-For")
	if len(fields) == 0 {
		return ""
	}
	parts := strings.Split(strings.Join(fields, ","), ",")

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
