package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/", nil)
	middleware.SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	csp := rec.Header().Get("Content-Security-Policy")
	require.Contains(t, csp, "script-src 'self';")
	assert.NotContains(t, csp, "script-src 'self' 'unsafe-inline'",
		"the Vue production build ships only external module scripts; inline script must not be allowed")
}

func TestCrossOriginProtection(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.CrossOriginProtection("")(next)

	t.Run("a safe method passes regardless of origin", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "https://example.com/api/me", nil)
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("a same-origin unsafe request passes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "https://example.com/api/projects", nil)
		req.Header.Set("Sec-Fetch-Site", "same-origin")
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("a cross-site unsafe request is rejected with a JSON 403", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "https://example.com/api/projects", nil)
		req.Header.Set("Sec-Fetch-Site", "cross-site")
		h.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "json")

		var body struct {
			Message string `json:"message"`
		}
		require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
		assert.NotEmpty(t, body.Message, "the deny response must carry a JSON message the SPA can surface")
	})

	t.Run("a same-site (sibling subdomain) unsafe request is rejected", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "https://example.com/api/projects", nil)
		req.Header.Set("Sec-Fetch-Site", "same-site")
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("a non-browser request that sends neither Sec-Fetch-Site nor Origin passes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "https://example.com/api/projects", nil)
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("a cross-origin Origin header without Sec-Fetch-Site is rejected", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "https://example.com/api/projects", nil)
		req.Host = "example.com"
		req.Header.Set("Origin", "https://evil.example")
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("a configured trusted origin is allowed through the Origin fallback", func(t *testing.T) {
		trusted := middleware.CrossOriginProtection("https://app.example.com")(next)

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "https://api.example.com/api/projects", nil)
		req.Host = "api.example.com"
		req.Header.Set("Origin", "https://app.example.com")
		trusted.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("an invalid trusted origin panics at construction", func(t *testing.T) {
		assert.Panics(t, func() {
			middleware.CrossOriginProtection("not-a-valid-origin")
		})
	})
}
