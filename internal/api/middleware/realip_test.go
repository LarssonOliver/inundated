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

	got := seenRemoteAddr(middleware.RealIP(nil), req)

	assert.Equal(t, "198.51.100.9:5000", got)
}

func TestRealIP_UntrustedPeerIgnoresForwardedHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "198.51.100.9:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8")), req)

	assert.Equal(t, "198.51.100.9:5000", got)
}

func TestRealIP_TrustedPeerUsesXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8")), req)

	assert.Equal(t, "203.0.113.7", got)
}

func TestRealIP_PicksRightmostUntrustedInChain(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	// client -> external proxy -> our ingress. Only the internal hop is trusted.
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 198.51.100.4, 10.9.9.9")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8")), req)

	assert.Equal(t, "198.51.100.4", got)
}

func TestRealIP_AllEntriesTrustedFallsBackToLeftmost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Forwarded-For", "10.4.4.4, 10.9.9.9")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8")), req)

	assert.Equal(t, "10.4.4.4", got)
}

func TestRealIP_FallsBackToXRealIPWhenNoXForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Real-IP", "203.0.113.7")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8")), req)

	assert.Equal(t, "203.0.113.7", got)
}

func TestRealIP_SkipsMalformedForwardedEntries(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"
	req.Header.Set("X-Forwarded-For", "203.0.113.7, garbage, 10.9.9.9")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8")), req)

	assert.Equal(t, "203.0.113.7", got)
}

func TestRealIP_TrustedPeerWithNoForwardedHeadersKeepsPeer(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "10.1.2.3:5000"

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "10.0.0.0/8")), req)

	assert.Equal(t, "10.1.2.3:5000", got)
}

func TestRealIP_IPv6TrustedPeerAndClient(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req.RemoteAddr = "[fd00::1]:5000"
	req.Header.Set("X-Forwarded-For", "2001:db8::1234")

	got := seenRemoteAddr(middleware.RealIP(mustPrefixes(t, "fd00::/8")), req)

	assert.Equal(t, "2001:db8::1234", got)
}
