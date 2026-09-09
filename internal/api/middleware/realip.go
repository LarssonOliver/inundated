package middleware

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

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
