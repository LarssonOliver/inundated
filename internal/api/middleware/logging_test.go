package middleware_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureDefaultLogger swaps slog's default logger for a debug-level JSON logger
// writing to the returned buffer, restoring the previous default on cleanup.
func captureDefaultLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(&logging.ContextHandler{
		Handler: slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
	}))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func lastRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	require.NotEmpty(t, lines)
	var rec map[string]any
	require.NoError(t, json.Unmarshal(lines[len(lines)-1], &rec))
	return rec
}

func TestRequestLoggerEmitsStructuredLine(t *testing.T) {
	buf := captureDefaultLogger(t)

	h := middleware.RequestLogger(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hi"))
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/projects", nil))

	rec := lastRecord(t, buf)
	assert.Equal(t, "http request", rec["msg"])
	assert.Equal(t, "GET", rec["method"])
	assert.Equal(t, "/api/projects", rec["path"])
	assert.Equal(t, float64(http.StatusOK), rec["status"])
	assert.Equal(t, float64(2), rec["bytes"])
	assert.Equal(t, "INFO", rec["level"])
}

func TestRequestLoggerLogsServerErrorsAtErrorLevel(t *testing.T) {
	buf := captureDefaultLogger(t)

	h := middleware.RequestLogger(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/projects", nil))

	assert.Equal(t, "ERROR", lastRecord(t, buf)["level"])
}

func TestRequestLoggerObservesPanicRecoveredAsServerError(t *testing.T) {
	buf := captureDefaultLogger(t)

	// RequestLogger must wrap Recoverer so a recovered panic is still logged
	// as a 500. Registration order: RequestLogger first (outer), Recoverer second.
	panicky := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	h := middleware.RequestLogger(nil)(chimiddleware.Recoverer(panicky))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/projects", nil))

	rec := lastRecord(t, buf)
	assert.Equal(t, "http request", rec["msg"])
	assert.Equal(t, "ERROR", rec["level"])
	assert.Equal(t, float64(http.StatusInternalServerError), rec["status"])
}

func TestRequestLoggerRespectsSkipFn(t *testing.T) {
	buf := captureDefaultLogger(t)

	skip := func(r *http.Request) bool { return r.URL.Path == "/health" }
	h := middleware.RequestLogger(skip)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/health", nil))

	assert.Empty(t, buf.String())
}

func TestRequestLogContextPropagatesRequestID(t *testing.T) {
	buf := captureDefaultLogger(t)

	inner := middleware.RequestLogger(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	h := chimiddleware.RequestID(middleware.RequestLogContext(inner))
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/api/projects", nil))

	rec := lastRecord(t, buf)
	assert.NotEmpty(t, rec["request_id"])
}

func TestRequestLogContextIsANoOpWithoutARequestID(t *testing.T) {
	// No chimiddleware.RequestID upstream: the middleware must not panic or inject a key.
	buf := captureDefaultLogger(t)

	inner := middleware.RequestLogger(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Nil(t, r.Context().Value(chimiddleware.RequestIDKey))
		w.WriteHeader(http.StatusOK)
	}))
	middleware.RequestLogContext(inner).ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/api/projects", nil).WithContext(context.Background()),
	)

	_, hasRequestID := lastRecord(t, buf)["request_id"]
	assert.False(t, hasRequestID)
}
