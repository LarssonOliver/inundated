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
	h := middleware.CSRF(key)(next)

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
}
