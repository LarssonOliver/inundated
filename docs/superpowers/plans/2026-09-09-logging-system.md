# Logging System Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the ad-hoc `log` package usage with a `log/slog`-based logging system that supports runtime level filtering, text/JSON output, and automatic request/user correlation.

**Architecture:** A new stdlib-only `internal/logging` package builds a `*slog.Logger` whose handler filters by level, formats as text or JSON, and flushes context-attached attributes onto every record. `main` builds it from config and installs it with `slog.SetDefault`, so all call sites use the package-level `slog.*` functions. Middleware seeds `request_id` (and, after auth, `user_id`) onto the request context; the custom handler picks them up.

**Tech Stack:** Go 1.25, `log/slog` (standard library), `github.com/go-chi/chi/v5` middleware, `testify`.

**Spec:** `docs/superpowers/specs/2026-09-09-logging-system-design.md`

## Global Constraints

- Go version floor: `go 1.25.5` (from `go.mod`) — `log/slog` and `slog.DiscardHandler` era APIs are available.
- No new direct dependency in `go.mod`. `log/slog` is standard library.
- `internal/logging` imports standard library only (no project packages), so any package may import it without a cycle.
- Flag/env convention (from `internal/config/config.go`): flag name is `lowercase-hyphen`, env var is `UPPER_SNAKE_CASE`, help text ends with `(env: VAR_NAME)`.
- All application logs go to **stdout**. Pre-logger failures and CLI `-help` output stay on **stderr**.
- Log levels are exactly `debug|info|warn|error`. Log formats are exactly `text|json`.
- End every commit message with the trailer block:
  ```
  Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>
  Claude-Session: https://claude.ai/code/session_01LsZAQAzh9pADLVZwWU7JnA
  ```
- Work on a branch, not `main`.

---

## File Structure

**Created:**
- `internal/logging/logging.go` — `Options`, `New`, `parseLevel`.
- `internal/logging/context.go` — `ContextHandler`, `ContextWith`, `attrsFromContext`.
- `internal/logging/logging_test.go` — black-box tests for the package.
- `internal/api/middleware/logging_test.go` — tests for `RequestLogger` and `RequestLogContext`.

**Modified:**
- `internal/config/config.go` — add `LogFormat` field, flag, validation.
- `internal/config/config_test.go` — tests for `LogFormat`.
- `internal/api/middleware/logging.go` — drop `*log.Logger` param from `RequestLogger`; add `RequestLogContext`; emit structured `slog` record with status-based level.
- `internal/api/middleware/auth.go` — attach `user_id` to the log context after auth resolves.
- `internal/api/middleware/auth_test.go` — assert `user_id` reaches the downstream context.
- `cmd/server/main.go` — build the logger, `slog.SetDefault`, `fatal` helper, convert every `log.*` call, route `http.Server.ErrorLog` through slog, update `newRouter`.
- `cmd/server/main_test.go` — assert `newHTTPServer` wires `ErrorLog`.
- `internal/api/handlers/auth.go` — `slog.WarnContext` for the OIDC callback failure.
- `internal/service/user.go` — `slog.InfoContext` for the orphan-adoption line.
- `internal/service/cleanup.go` — `slog.ErrorContext` for sweep failures, `slog.DebugContext` for a completed pass.
- `internal/service/cleanup_test.go` — capture `slog` output instead of `log`.
- `README.md` — document `LOG_LEVEL` and `LOG_FORMAT`.

---

## Task 1: `internal/logging` package

**Files:**
- Create: `internal/logging/logging.go`
- Create: `internal/logging/context.go`
- Test: `internal/logging/logging_test.go`

**Interfaces:**
- Consumes: nothing (stdlib only).
- Produces:
  - `logging.Options` struct: `Level string`, `Format string`, `Writer io.Writer`.
  - `logging.New(opts Options) (*slog.Logger, error)` — error on unknown level/format.
  - `logging.ContextHandler` struct with an embedded `slog.Handler` (exported field name `Handler`).
  - `logging.ContextWith(ctx context.Context, attrs ...slog.Attr) context.Context`.

- [ ] **Step 1: Write the failing tests**

Create `internal/logging/logging_test.go`:

```go
package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/larssonoliver/inundated/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFiltersBelowConfiguredLevel(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Level: "warn", Format: "json", Writer: &buf})
	require.NoError(t, err)

	l.Info("should be hidden")
	l.Warn("should be shown")

	out := buf.String()
	assert.NotContains(t, out, "should be hidden")
	assert.Contains(t, out, "should be shown")
}

func TestNewDefaultsToTextFormat(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Writer: &buf})
	require.NoError(t, err)

	l.Info("hello", "key", "value")

	assert.Contains(t, buf.String(), "key=value")
}

func TestNewJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)

	l.Info("hello", "key", "value")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "hello", rec["msg"])
	assert.Equal(t, "value", rec["key"])
}

func TestNewRejectsUnknownLevel(t *testing.T) {
	_, err := logging.New(logging.Options{Level: "verbose"})
	assert.Error(t, err)
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	_, err := logging.New(logging.Options{Format: "yaml"})
	assert.Error(t, err)
}

func TestContextAttrsAppearInEveryRecord(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)

	ctx := logging.ContextWith(context.Background(), slog.String("request_id", "abc123"))
	l.InfoContext(ctx, "hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "abc123", rec["request_id"])
}

func TestContextWithAccumulates(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)

	ctx := logging.ContextWith(context.Background(), slog.String("request_id", "abc"))
	ctx = logging.ContextWith(ctx, slog.String("user_id", "u-1"))
	l.InfoContext(ctx, "hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "abc", rec["request_id"])
	assert.Equal(t, "u-1", rec["user_id"])
}

func TestContextAttrsSurviveLoggerWith(t *testing.T) {
	var buf bytes.Buffer
	l, err := logging.New(logging.Options{Format: "json", Writer: &buf})
	require.NoError(t, err)

	ctx := logging.ContextWith(context.Background(), slog.String("request_id", "abc"))
	l.With("component", "test").InfoContext(ctx, "hello")

	var rec map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &rec))
	assert.Equal(t, "abc", rec["request_id"])
	assert.Equal(t, "test", rec["component"])
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/logging/ -v`
Expected: FAIL — build error, `logging.New`/`logging.Options`/`logging.ContextWith` undefined.

- [ ] **Step 3: Implement `context.go`**

Create `internal/logging/context.go`:

```go
package logging

import (
	"context"
	"log/slog"
)

type ctxKey struct{}

// ContextWith returns a context carrying attrs that ContextHandler adds to every
// record subsequently logged with a derived context. Repeated calls accumulate;
// a later attr with the same key shadows an earlier one per slog's last-wins rule.
func ContextWith(ctx context.Context, attrs ...slog.Attr) context.Context {
	if len(attrs) == 0 {
		return ctx
	}
	prev, _ := ctx.Value(ctxKey{}).([]slog.Attr)
	next := make([]slog.Attr, 0, len(prev)+len(attrs))
	next = append(next, prev...)
	next = append(next, attrs...)
	return context.WithValue(ctx, ctxKey{}, next)
}

func attrsFromContext(ctx context.Context) []slog.Attr {
	attrs, _ := ctx.Value(ctxKey{}).([]slog.Attr)
	return attrs
}

// ContextHandler wraps a base slog.Handler and, on each record, appends the
// attributes accumulated on the record's context by ContextWith.
type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := attrsFromContext(ctx); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{Handler: h.Handler.WithGroup(name)}
}
```

- [ ] **Step 4: Implement `logging.go`**

Create `internal/logging/logging.go`:

```go
package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

// Options configures the process logger. The zero value is valid: it produces a
// text logger at info level writing to stdout.
type Options struct {
	Level  string    // "debug"|"info"|"warn"|"error"; "" means "info"
	Format string    // "text"|"json"; "" means "text"
	Writer io.Writer // nil means os.Stdout
}

// New builds a *slog.Logger whose handler filters at the configured level,
// formats as text or JSON, and adds context attributes recorded by ContextWith.
// It returns an error for an unrecognized Level or Format.
func New(opts Options) (*slog.Logger, error) {
	level, err := parseLevel(opts.Level)
	if err != nil {
		return nil, err
	}

	w := opts.Writer
	if w == nil {
		w = os.Stdout
	}

	handlerOpts := &slog.HandlerOptions{Level: level}

	var base slog.Handler
	switch strings.ToLower(opts.Format) {
	case "", "text":
		base = slog.NewTextHandler(w, handlerOpts)
	case "json":
		base = slog.NewJSONHandler(w, handlerOpts)
	default:
		return nil, fmt.Errorf("logging: unknown format %q (want text or json)", opts.Format)
	}

	return slog.New(&ContextHandler{Handler: base}), nil
}

func parseLevel(name string) (slog.Level, error) {
	switch strings.ToLower(name) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("logging: unknown level %q (want debug, info, warn, or error)", name)
	}
}
```

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/logging/ -v`
Expected: PASS (all 8 tests).

- [ ] **Step 6: Vet and lint**

Run: `go vet ./internal/logging/ && golangci-lint run ./internal/logging/...`
Expected: no output / exit 0.

- [ ] **Step 7: Commit**

```bash
git add internal/logging/
git commit -m "feat(logging): slog-based logger with level, format, and context correlation"
```

---

## Task 2: `internal/config` — `LogFormat`

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: `config.Config.LogFormat string` — value is always `"text"` or `"json"` after a successful `Load`.

- [ ] **Step 1: Write the failing tests**

Add to `internal/config/config_test.go`:

```go
func TestLogFormatDefaultsToText(t *testing.T) {
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(nil)))
	assert.NoError(t, err)
	assert.Equal(t, "text", cfg.LogFormat)
}

func TestLogFormatFromEnv(t *testing.T) {
	env := map[string]string{"LOG_FORMAT": "json"}
	cfg, err := config.Load(config.WithArgs(nil), config.WithEnvLookup(fakeEnv(env)))
	assert.NoError(t, err)
	assert.Equal(t, "json", cfg.LogFormat)
}

func TestLogFormatFromFlag(t *testing.T) {
	cfg, err := config.Load(config.WithArgs([]string{"-log-format=json"}), config.WithEnvLookup(fakeEnv(nil)))
	assert.NoError(t, err)
	assert.Equal(t, "json", cfg.LogFormat)
}

func TestValidationRejectsInvalidLogFormat(t *testing.T) {
	_, err := config.Load(config.WithArgs([]string{"-log-format=yaml"}), config.WithEnvLookup(fakeEnv(nil)))
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/config/ -run TestLogFormat -v`
Expected: FAIL — `cfg.LogFormat` undefined (build error).

- [ ] **Step 3: Add the struct field**

In `internal/config/config.go`, in `type Config struct`, add `LogFormat` right after `LogLevel`:

```go
	DatabaseURL             string
	LogLevel                string
	LogFormat               string
	PublicBaseURL           string
```

- [ ] **Step 4: Register the flag**

In `func (l *loader) load()`, right after the `logLevel := fs.String(...)` block:

```go
	logFormat := fs.String("log-format", l.envOr("LOG_FORMAT", "text"),
		"Log format: text|json (env: LOG_FORMAT)")
```

- [ ] **Step 5: Populate the field**

In the `cfg := &Config{...}` literal, right after `LogLevel: *logLevel,`:

```go
		LogFormat:     *logFormat,
```

- [ ] **Step 6: Validate**

In `func (c *Config) validate()`, right after the `validLevels` check block:

```go
	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[c.LogFormat] {
		return fmt.Errorf("config: log-format %q is not one of text|json", c.LogFormat)
	}
```

- [ ] **Step 7: Run the tests to verify they pass**

Run: `go test ./internal/config/ -v`
Expected: PASS (new tests plus all existing config tests).

- [ ] **Step 8: Commit**

```bash
git add internal/config/
git commit -m "feat(config): add LOG_FORMAT option (text|json)"
```

---

## Task 3: `RequestLogger` refactor + `RequestLogContext`

**Files:**
- Modify: `internal/api/middleware/logging.go`
- Modify: `cmd/server/main.go:74-120` (`newRouter` only)
- Test: `internal/api/middleware/logging_test.go` (create)

**Interfaces:**
- Consumes: `logging.ContextWith` (Task 1); `chimiddleware.GetReqID` from `github.com/go-chi/chi/v5/middleware`.
- Produces:
  - `middleware.RequestLogger(skipFn func(r *http.Request) bool) func(http.Handler) http.Handler` — **note the dropped `*log.Logger` parameter**.
  - `middleware.RequestLogContext(next http.Handler) http.Handler` — a plain `func(http.Handler) http.Handler`-compatible middleware.

- [ ] **Step 1: Write the failing tests**

Create `internal/api/middleware/logging_test.go`:

```go
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
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/api/middleware/ -run 'RequestLog' -v`
Expected: FAIL — `RequestLogger` signature mismatch and `RequestLogContext` undefined.

- [ ] **Step 3: Rewrite `internal/api/middleware/logging.go`**

Replace the whole file with:

```go
package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/larssonoliver/inundated/internal/logging"
)

// RequestLogContext seeds the request's chi RequestID onto the logging context
// so every record emitted while handling the request carries request_id.
// Place it immediately after chimiddleware.RequestID.
func RequestLogContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := middleware.GetReqID(r.Context()); id != "" {
			r = r.WithContext(logging.ContextWith(r.Context(), slog.String("request_id", id)))
		}
		next.ServeHTTP(w, r)
	})
}

// RequestLogger emits one structured record per request via slog's default
// logger. skipFn, when non-nil and true for a request, suppresses the record.
func RequestLogger(skipFn func(r *http.Request) bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if skipFn != nil && skipFn(r) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			level := slog.LevelInfo
			if ww.Status() >= http.StatusInternalServerError {
				level = slog.LevelError
			}

			slog.LogAttrs(r.Context(), level, "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", ww.Status()),
				slog.Int("bytes", ww.BytesWritten()),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}
```

- [ ] **Step 4: Update `newRouter` in `cmd/server/main.go`**

In `newRouter`, add `RequestLogContext` right after the `RequestID` line:

```go
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.RequestLogContext)
	r.Use(middleware.RealIP(cfg.TrustedProxies, cfg.TrustedProxyHeaders))
```

In the `r.Group(...)` block, delete the `logger := log.New(...)` line and drop the argument:

```go
	r.Group(func(r chi.Router) {
		r.Use(middleware.RequestLogger(func(r *http.Request) bool {
			return r.URL.Path == "/health"
		}))
```

Leave every other `log.*` call in `main.go` untouched for now (Task 5 handles them), so the `"log"` and `"os"` imports stay.

- [ ] **Step 5: Run the tests to verify they pass**

Run: `go test ./internal/api/middleware/ ./cmd/server/ -v`
Expected: PASS — new `RequestLog*` tests pass; existing middleware and `main` tests still pass.

- [ ] **Step 6: Vet and lint**

Run: `go vet ./... && golangci-lint run`
Expected: exit 0.

- [ ] **Step 7: Commit**

```bash
git add internal/api/middleware/logging.go internal/api/middleware/logging_test.go cmd/server/main.go
git commit -m "feat(middleware): structured request logging with request-id correlation"
```

---

## Task 4: Attach `user_id` to the log context in `OIDCAuth`

**Files:**
- Modify: `internal/api/middleware/auth.go:96-99`
- Test: `internal/api/middleware/auth_test.go` (extend the existing "Valid session - attaches context successfully" case)

**Interfaces:**
- Consumes: `logging.ContextWith` (Task 1); `model.User.Id` is a `uuid.UUID` (has `.String()`).
- Produces: no new symbols. After `OIDCAuth` authenticates a request, its downstream context carries `slog.String("user_id", <uuid>)`.

- [ ] **Step 1: Extend the failing test**

In `internal/api/middleware/auth_test.go`, add these imports:

```go
	"bytes"

	"github.com/larssonoliver/inundated/internal/logging"
```

In the test case named `"Valid session - attaches context successfully"`, append to its `checkResult` body (after the existing session assertions):

```go
				var buf bytes.Buffer
				lg, err := logging.New(logging.Options{Format: "json", Writer: &buf})
				require.NoError(t, err)
				lg.InfoContext(lastSeenCtx, "probe")
				assert.Contains(t, buf.String(), userID.String(),
					"an authenticated request's context must carry user_id for logging")
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/api/middleware/ -run 'TestOIDCAuth/Valid_session_-_attaches_context_successfully' -v`
Expected: FAIL — buffer does not contain the user ID.

- [ ] **Step 3: Implement**

In `internal/api/middleware/auth.go`, add imports:

```go
	"log/slog"

	"github.com/larssonoliver/inundated/internal/logging"
```

Change the context-building lines near the end of `OIDCAuth` from:

```go
			ctx := model.SetSessionInContext(r.Context(), session)
			ctx = model.SetUserInContext(ctx, user)

			next.ServeHTTP(w, r.WithContext(ctx))
```

to:

```go
			ctx := model.SetSessionInContext(r.Context(), session)
			ctx = model.SetUserInContext(ctx, user)
			ctx = logging.ContextWith(ctx, slog.String("user_id", user.Id.String()))

			next.ServeHTTP(w, r.WithContext(ctx))
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `go test ./internal/api/middleware/ -v`
Expected: PASS (all `TestOIDCAuth` subtests).

- [ ] **Step 5: Commit**

```bash
git add internal/api/middleware/auth.go internal/api/middleware/auth_test.go
git commit -m "feat(middleware): tag authenticated request logs with user_id"
```

---

## Task 5: Convert `cmd/server/main.go` to slog

**Files:**
- Modify: `cmd/server/main.go`
- Test: `cmd/server/main_test.go`

**Interfaces:**
- Consumes: `logging.New`, `config.Config.LogLevel`, `config.Config.LogFormat`.
- Produces: `fatal(msg string, args ...any)` package-private helper in `main` (logs at error, then `os.Exit(1)`).

- [ ] **Step 1: Write the failing test**

Add to `cmd/server/main_test.go`:

```go
func TestNewHTTPServer_RoutesErrorsThroughSlog(t *testing.T) {
	s := newHTTPServer("127.0.0.1:0", http.NotFoundHandler())
	assert.NotNil(t, s.ErrorLog, "the stdlib server's internal errors must be routed through slog")
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./cmd/server/ -run TestNewHTTPServer_RoutesErrorsThroughSlog -v`
Expected: FAIL — `s.ErrorLog` is nil.

- [ ] **Step 3: Update imports**

In `cmd/server/main.go`, remove `"log"` and add `"log/slog"`, and add the package import:

```go
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/larssonoliver/inundated/internal/api"
	"github.com/larssonoliver/inundated/internal/api/handlers"
	"github.com/larssonoliver/inundated/internal/api/middleware"
	"github.com/larssonoliver/inundated/internal/auth"
	"github.com/larssonoliver/inundated/internal/config"
	postgresdb "github.com/larssonoliver/inundated/internal/db/postgres"
	"github.com/larssonoliver/inundated/internal/logging"
	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/repository/memory"
	"github.com/larssonoliver/inundated/internal/repository/postgres"
	"github.com/larssonoliver/inundated/internal/service"
```

- [ ] **Step 4: Add the `fatal` helper**

Add near the top of `cmd/server/main.go` (e.g. right after `var Version = "dev"`):

```go
// fatal logs msg at error level with the given key/value args, then exits 1.
func fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
```

- [ ] **Step 5: Convert `setupRepositories`**

```go
func setupRepositories(ctx context.Context, databaseUrl string) (
	repository.Repository,
	repository.LoginStateRepository,
	repository.SessionRepository,
) {
	if databaseUrl == "in-memory" {
		slog.Warn("using in-memory repository", "note", "not recommended for production")
		memoryStore := memory.NewMemoryStore()
		return memoryStore, memoryStore, memoryStore
	}
	if strings.HasPrefix(databaseUrl, "postgresql://") {
		slog.Info("using postgres repository")
		slog.Info("applying database migrations")
		err := postgresdb.ApplyMigrations(ctx, databaseUrl)
		if err != nil {
			fatal("failed to apply database migrations", "error", err)
		}

		repository, err := postgres.NewPostgresStore(ctx, databaseUrl)
		if err != nil {
			fatal("failed to connect to postgresql", "error", err)
		}
		return repository, repository, repository
	}
	fatal("unsupported database url", "url", databaseUrl)
	return nil, nil, nil
}
```

- [ ] **Step 6: Build the logger in `main` and set it as default**

In `func main()`, right after the `cfg, err := config.Load()` error-handling block and before `ctx := context.Background()`:

```go
	logger, err := logging.New(logging.Options{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "logging setup: %v\n", err)
		os.Exit(1)
	}
	slog.SetDefault(logger)
```

- [ ] **Step 7: Convert the remaining `log.*` calls in `main`**

```go
	if err := service.EnsureAuthConfigConsistent(ctx, repo, cfg.OIDC.Enabled()); err != nil {
		fatal("auth configuration", "error", err)
	}

	if cfg.DisableUserRegistration {
		slog.Info("user registration disabled; only OIDC identities with an existing account can log in")
		if hasUsers, err := repo.HasUsers(ctx); err == nil && !hasUsers {
			slog.Warn("registration is disabled and no users exist yet -- nobody can log in until you enable it once")
		}
	}
```

```go
	if cfg.OIDC.Enabled() {
		slog.Info("OIDC authentication enabled", "issuer", cfg.OIDC.IssuerURL)
	} else {
		slog.Info("OIDC not configured; running in userless mode")
	}

	if !shouldUseSecureCookies(cfg) {
		slog.Warn("insecure cookies (Secure attribute off); set PUBLIC_BASE_URL to an https:// origin in production")
	}

	if len(cfg.TrustedProxies) == 0 {
		slog.Warn("TRUSTED_PROXIES not set; forwarded-for headers are ignored and rate " +
			"limiting keys off the direct connection address. Behind a reverse proxy that " +
			"means every client shares one bucket -- set TRUSTED_PROXIES to the proxy's address(es)")
	}

	slog.Info("starting inundated", "version", Version, "addr", addrStr)
	if err := s.ListenAndServe(); err != nil {
		fatal("http server stopped", "error", err)
	}
```

- [ ] **Step 8: Route the stdlib server's errors through slog**

In `newHTTPServer`, add the `ErrorLog` field:

```go
func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ErrorLog:          slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
```

- [ ] **Step 9: Run the full test suite**

Run: `go build ./... && go test ./... `
Expected: PASS. `main.go` no longer imports `"log"`; `go build` confirms no stale references.

- [ ] **Step 10: Vet and lint**

Run: `go vet ./... && golangci-lint run`
Expected: exit 0.

- [ ] **Step 11: Commit**

```bash
git add cmd/server/main.go cmd/server/main_test.go
git commit -m "feat(server): build the slog logger from config and convert startup logging"
```

---

## Task 6: Convert the leaf call sites

**Files:**
- Modify: `internal/api/handlers/auth.go`
- Modify: `internal/service/user.go`
- Modify: `internal/service/cleanup.go`
- Test: `internal/service/cleanup_test.go`

**Interfaces:**
- Consumes: `slog` package-level functions against the default logger installed in Task 5.
- Produces: no new symbols.

- [ ] **Step 1: Rewrite the cleanup tests to capture slog**

Replace `internal/service/cleanup_test.go` with:

```go
package service_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/larssonoliver/inundated/internal/repository"
	"github.com/larssonoliver/inundated/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureInfoLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestCleanupService_LogsSweepErrorsAndKeepsGoing(t *testing.T) {
	buf := captureInfoLogs(t)

	var sessionsCalled, loginStatesCalled bool
	sessions := &repository.SessionRepoMock{
		DeleteAllExpiredSessionsFn: func(context.Context) error {
			sessionsCalled = true
			return errors.New("sessions sweep failed")
		},
	}
	loginStates := &repository.LoginStateRepoMock{
		DeleteAllExpiredLoginStatesFn: func(context.Context) error {
			loginStatesCalled = true
			return errors.New("login-states sweep failed")
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service.NewCleanupService(sessions, loginStates, time.Hour).Run(ctx)

	assert.True(t, sessionsCalled, "the session sweep must run")
	assert.True(t, loginStatesCalled, "a failed session sweep must not skip the login-state sweep")

	out := buf.String()
	require.Contains(t, out, "sessions sweep failed")
	require.Contains(t, out, "login-states sweep failed")
	require.Contains(t, out, `"level":"ERROR"`)
}

func TestCleanupService_QuietWhenSweepsSucceed(t *testing.T) {
	buf := captureInfoLogs(t)

	sessions := &repository.SessionRepoMock{
		DeleteAllExpiredSessionsFn: func(context.Context) error { return nil },
	}
	loginStates := &repository.LoginStateRepoMock{
		DeleteAllExpiredLoginStatesFn: func(context.Context) error { return nil },
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service.NewCleanupService(sessions, loginStates, time.Hour).Run(ctx)

	assert.Empty(t, buf.String(), "a clean sweep must not log at info level or above")
}

func TestCleanupService_EmitsDebugOnCompletedPass(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	sessions := &repository.SessionRepoMock{
		DeleteAllExpiredSessionsFn: func(context.Context) error { return nil },
	}
	loginStates := &repository.LoginStateRepoMock{
		DeleteAllExpiredLoginStatesFn: func(context.Context) error { return nil },
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service.NewCleanupService(sessions, loginStates, time.Hour).Run(ctx)

	var rec map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &rec))
	assert.Equal(t, "cleanup run complete", rec["msg"])
	assert.Equal(t, "DEBUG", rec["level"])
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `go test ./internal/service/ -run TestCleanupService -v`
Expected: FAIL — `TestCleanupService_EmitsDebugOnCompletedPass` fails (no debug line); the error-level assertion fails (still using `log`).

- [ ] **Step 3: Convert `internal/service/cleanup.go`**

Replace the import block's `"log"` with `"log/slog"`, and rewrite `cleanup`:

```go
func (c *CleanupServiceImpl) cleanup(ctx context.Context) {
	if err := c.sessionRepository.DeleteAllExpiredSessions(ctx); err != nil {
		slog.ErrorContext(ctx, "cleanup: deleting expired sessions failed", "error", err)
	}
	if err := c.loginStateRepository.DeleteAllExpiredLoginStates(ctx); err != nil {
		slog.ErrorContext(ctx, "cleanup: deleting expired login states failed", "error", err)
	}
	slog.DebugContext(ctx, "cleanup run complete")
}
```

- [ ] **Step 4: Convert `internal/api/handlers/auth.go`**

Replace the import `"log"` with `"log/slog"`. Change line 108:

```go
	session, token, redirectUrl, err := a.svc.HandleCallback(ctx, stateId, request.Params.Code)
	if err != nil {
		slog.WarnContext(ctx, "OIDC callback failed", "error", err)
		return api.AuthCallback401Response{}, nil
	}
```

- [ ] **Step 5: Convert `internal/service/user.go`**

Replace the import `"log"` with `"log/slog"`. Change the adoption log (lines ~78-83):

```go
	if adoption.Total() > 0 {
		slog.InfoContext(ctx, "first user adopted orphaned resources",
			"user_id", created.Id,
			"projects", adoption.Projects,
			"tags", adoption.Tags,
			"timespans", adoption.Timespans,
		)
	}
```

- [ ] **Step 6: Run the full test suite**

Run: `go build ./... && go test ./...`
Expected: PASS. No package imports `"log"` any more except test helpers that legitimately need it (there should be none left in non-test code).

- [ ] **Step 7: Confirm no stray `log` usage remains**

Run: `grep -rn '"log"' --include='*.go' internal/ cmd/ | grep -v _test`
Expected: no output.

- [ ] **Step 8: Vet and lint**

Run: `go vet ./... && golangci-lint run`
Expected: exit 0.

- [ ] **Step 9: Commit**

```bash
git add internal/api/handlers/auth.go internal/service/user.go internal/service/cleanup.go internal/service/cleanup_test.go
git commit -m "refactor: route handler and service logging through slog"
```

---

## Task 7: Document the logging options

**Files:**
- Modify: `README.md`

**Interfaces:**
- Consumes: nothing.
- Produces: nothing.

- [ ] **Step 1: Locate the configuration section**

Run: `grep -n 'LOG_LEVEL\|DATABASE_URL\|Configuration\|| Env\|env:' README.md`
Expected: finds the configuration/environment-variable table or list.

- [ ] **Step 2: Add / update entries**

In the configuration reference, ensure both of these are documented (match the surrounding table or list style):

- `LOG_LEVEL` — `debug|info|warn|error` (default `info`). Controls which records are emitted. *(This env var already existed but was previously inert; note that it is now wired.)*
- `LOG_FORMAT` — `text|json` (default `text`). `text` is human-readable for local use; set `json` in production for log aggregators. All logs are written to stdout.

- [ ] **Step 3: Verify the doc builds / renders**

Run: `grep -n 'LOG_FORMAT' README.md`
Expected: the new entry is present.

- [ ] **Step 4: Commit**

```bash
git add README.md
git commit -m "docs: document LOG_LEVEL and LOG_FORMAT"
```

---

## Final Verification

- [ ] **Step 1: Full build and test**

Run: `go build ./... && go test ./...`
Expected: PASS across every package.

- [ ] **Step 2: Lint**

Run: `golangci-lint run`
Expected: exit 0.

- [ ] **Step 3: Manual smoke — text**

Run: `go run ./cmd/server 2>&1 | head -5` (Ctrl-C after a few lines)
Expected: human-readable `time=... level=INFO msg="using in-memory repository" ...` and `msg="starting inundated" version=dev addr=0.0.0.0:8080`.

- [ ] **Step 4: Manual smoke — json + level**

Run: `LOG_FORMAT=json LOG_LEVEL=warn go run ./cmd/server 2>&1 | head -5`
Expected: JSON lines only; the `INFO` "starting inundated" line is filtered out, the `WARN` in-memory-repo line is present.

- [ ] **Step 5: Manual smoke — request correlation**

With the server running (`LOG_FORMAT=json go run ./cmd/server`), in another shell: `curl -s localhost:8080/api/projects >/dev/null`
Expected: one `"msg":"http request"` line with `method`, `path`, `status`, `bytes`, `duration`, and a non-empty `request_id`.

- [ ] **Step 6: Update the branch and open a PR** (only if the user asks)

---

## Self-Review Notes

- **Spec coverage:** `internal/logging` package → Task 1. `LogFormat` config → Task 2. Global default logger / `SetDefault` → Task 5. `ContextHandler` + correlation → Tasks 1, 3, 4. Output format text/json, stdout → Tasks 1, 5. Level table for `main` → Task 5. Leaf call sites → Task 6. `RequestLogger` signature change + `RequestLogContext` → Task 3. `http.Server.ErrorLog` → Task 5. Tests for each package → every task. README → Task 7. No gaps.
- **Type consistency:** `logging.New(logging.Options) (*slog.Logger, error)`, `logging.ContextWith(context.Context, ...slog.Attr) context.Context`, `logging.ContextHandler{Handler: ...}`, `middleware.RequestLogger(func(*http.Request) bool)`, `middleware.RequestLogContext(http.Handler) http.Handler`, `fatal(string, ...any)` — used identically everywhere they appear.
- **Ordering:** each task leaves the tree building and green. Task 3 changes `RequestLogger`'s signature and its only caller (`newRouter`) in the same task. Task 5 removes the `"log"` import from `main.go` only after Task 3 has already removed `main.go`'s last non-`main()` use of it; the `main()` uses are converted in the same task (5).
