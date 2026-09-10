package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitByIP(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := middleware.RateLimitByIP(3, time.Minute)(ok)

	call := func(remoteAddr string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
		req.RemoteAddr = remoteAddr
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	t.Run("requests within the budget pass", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			assert.Equal(t, http.StatusOK, call("10.0.0.1:1111"))
		}
	})

	t.Run("the next request from the same IP is a JSON 429", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
		req.RemoteAddr = "10.0.0.1:1111"
		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.Contains(t, rec.Header().Get("Content-Type"), "json")
		assert.Contains(t, rec.Body.String(), "rate limit exceeded")
	})

	t.Run("a different IP has its own budget", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, call("10.0.0.2:2222"))
	})
}

func TestRateLimitByIPLogsDroppedRequests(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := middleware.RateLimitByIP(1, time.Minute)(ok)

	call := func() {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
		req.RemoteAddr = "10.9.0.1:1111"
		h.ServeHTTP(rec, req)
	}

	call()
	assert.Empty(t, buf.String(), "a request within budget must not log")

	call()
	out := buf.String()
	assert.Contains(t, out, "rate limit exceeded; dropping request")
	assert.Contains(t, out, "client_ip=10.9.0.1")
	assert.Contains(t, out, "method=GET")
	assert.Contains(t, out, "path=/api/projects")
}

func TestRateLimitByIPForPrefixes(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := middleware.RateLimitByIPForPrefixes(2, time.Minute, "/api/auth/")(ok)

	call := func(path string) int {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.RemoteAddr = "10.1.0.1:1111"
		h.ServeHTTP(rec, req)
		return rec.Code
	}

	t.Run("matching paths share the tight budget", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, call("/api/auth/login"))
		assert.Equal(t, http.StatusOK, call("/api/auth/callback"))
		assert.Equal(t, http.StatusTooManyRequests, call("/api/auth/login"))
	})

	t.Run("non-matching paths are never limited by this middleware", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			assert.Equal(t, http.StatusOK, call("/api/projects"))
		}
	})
}
