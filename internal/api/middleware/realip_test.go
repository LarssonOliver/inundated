package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/larssonoliver/inundated/internal/api/middleware"
)

var defaultForwardHeaders = []string{"X-Forwarded-For"}

func mustPrefixes(t *testing.T, ss ...string) []netip.Prefix {
	t.Helper()
	out := make([]netip.Prefix, len(ss))
	for i, s := range ss {
		p, err := netip.ParsePrefix(s)
		require.NoError(t, err)
		out[i] = p
	}
	return out
}

// seenRemoteAddr runs the middleware and reports the RemoteAddr the wrapped
// handler observed.
func seenRemoteAddr(h func(http.Handler) http.Handler, req *http.Request) string {
	var seen string
	h(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = r.RemoteAddr
	})).ServeHTTP(httptest.NewRecorder(), req)
	return seen
}

func TestRealIP_NoTrustedProxiesIgnoresForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "198.51.100.9:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")
	req.Header.Set("X-Real-IP", "203.0.113.8")

	got := seenRemoteAddr(middleware.RealIP(nil, defaultForwardHeaders), req)

	assert.Equal(t, "198.51.100.9:5000", got)
}

func TestRealIP_UntrustedPeerIgnoresForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "198.51.100.9:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), defaultForwardHeaders), req)

	assert.Equal(t, "198.51.100.9:5000", got)
}

func TestRealIP_TrustedPeerUsesXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), defaultForwardHeaders), req)

	assert.Equal(t, "203.0.113.7", got)
}

func TestRealIP_PicksRightmostUntrustedInChain(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	// client -> external proxy -> our ingress. Only the internal hop is trusted.
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 198.51.100.4, 10.9.9.9")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), defaultForwardHeaders), req)

	assert.Equal(t, "198.51.100.4", got)
}

func TestRealIP_AllEntriesTrustedFallsBackToLeftmost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Forwarded-For", "10.4.4.4, 10.9.9.9")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), defaultForwardHeaders), req)

	assert.Equal(t, "10.4.4.4", got)
}

func TestRealIP_JoinsMultipleXForwardedForHeaderLines(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	// A client-supplied line arrives first; the proxy appends its own as a
	// separate header field rather than editing the first.
	req.Header.Add("X-Forwarded-For", "1.2.3.4")
	req.Header.Add("X-Forwarded-For", "203.0.113.9")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), defaultForwardHeaders), req)

	assert.Equal(t, "203.0.113.9", got, "the spoofed leading field must not win")
}

func TestRealIP_SkipsMalformedForwardedEntries(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.7, garbage, 10.9.9.9")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), defaultForwardHeaders), req)

	assert.Equal(t, "203.0.113.7", got)
}

func TestRealIP_TrustedPeerWithNoForwardedHeadersKeepsPeer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), defaultForwardHeaders), req)

	assert.Equal(t, "10.1.2.3:5000", got)
}

func TestRealIP_IPv6TrustedPeerAndClient(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "[fd00::1]:5000"
	req.Header.Set("X-Forwarded-For", "2001:db8::1234")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "fd00::/8"), defaultForwardHeaders), req)

	assert.Equal(t, "2001:db8::1234", got)
}

func TestRealIP_XRealIPIgnoredWhenNotConfigured(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Real-IP", "203.0.113.7")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), defaultForwardHeaders), req)

	assert.Equal(t, "10.1.2.3:5000", got, "X-Real-IP must be ignored unless it is in the configured header list")
}

func TestRealIP_XRealIPHonoredWhenConfigured(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Real-IP", "203.0.113.7")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), []string{"X-Forwarded-For", "X-Real-IP"}), req)

	assert.Equal(t, "203.0.113.7", got)
}

func TestRealIP_EmptyHeaderListDefaultsToXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), nil), req)

	assert.Equal(t, "203.0.113.7", got)
}

func TestRealIP_HeaderListIsPriorityOrdered(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Real-IP", "203.0.113.7")
	req.Header.Set("X-Forwarded-For", "198.51.100.4")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8"), []string{"X-Real-IP", "X-Forwarded-For"}), req)

	assert.Equal(t, "203.0.113.7", got, "the first configured header that yields an address wins")
}
