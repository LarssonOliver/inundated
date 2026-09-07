package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCSRF(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.CSRF(key, true)(middleware.ExposeCSRFToken(true)(next))

	t.Run("safe method passes and receives a token cookie", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "https://example.com/api/me", nil)
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		require.NotEmpty(t, rec.Result().Cookies(), "expected a CSRF cookie to be set")
	})

	t.Run("unsafe method without a token is rejected with 403", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "https://example.com/api/projects", nil)
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("SPA round trip: the readable XSRF-TOKEN cookie authorizes an unsafe request", func(t *testing.T) {
		getRec := httptest.NewRecorder()
		getReq := httptest.NewRequest(http.MethodGet, "https://example.com/api/me", nil)
		h.ServeHTTP(getRec, getReq)

		cookies := getRec.Result().Cookies()
		var token string
		for _, c := range cookies {
			if c.Name == "XSRF-TOKEN" {
				token = c.Value
			}
		}
		require.NotEmpty(t, token, "expected a JS-readable XSRF-TOKEN cookie on a safe response")

		postRec := httptest.NewRecorder()
		postReq := httptest.NewRequest(http.MethodPost, "https://example.com/api/projects", nil)
		for _, c := range cookies {
			postReq.AddCookie(c)
		}
		postReq.Header.Set("X-XSRF-TOKEN", token)
		postReq.Header.Set("Origin", "https://example.com")
		h.ServeHTTP(postRec, postReq)

		assert.Equal(t, http.StatusOK, postRec.Code)
	})

	t.Run("insecure mode omits the Secure attribute so the cookie survives plain HTTP", func(t *testing.T) {
		insecure := middleware.CSRF(key, false)(middleware.ExposeCSRFToken(false)(next))

		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api/me", nil)
		insecure.ServeHTTP(rec, req)

		for _, c := range rec.Result().Cookies() {
			assert.False(t, c.Secure, "cookie %q must not be Secure in insecure mode", c.Name)
		}
	})
}
