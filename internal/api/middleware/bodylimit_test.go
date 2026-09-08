package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// echoBodyLen reads the whole body and reports whether the read succeeded.
func echoBodyLen(t *testing.T) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			if _, err := io.ReadAll(r.Body); err != nil {
				http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestMaxBodyBytes(t *testing.T) {
	const limit = 1 << 10 // 1 KiB
	h := middleware.MaxBodyBytes(limit)(echoBodyLen(t))

	t.Run("a body within the limit passes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(strings.Repeat("a", limit)))
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("a body over the limit is rejected before the handler can buffer it", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/projects", strings.NewReader(strings.Repeat("a", limit+1)))
		h.ServeHTTP(rec, req)

		assert.NotEqual(t, http.StatusOK, rec.Code)
	})

	t.Run("a nil body does not panic", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
		req.Body = nil
		require.NotPanics(t, func() { h.ServeHTTP(rec, req) })
	})
}
