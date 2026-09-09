# Logging system with log levels

Issue: #13

## Problem

The app logs with the standard `log` package throughout:

- `cmd/server/main.go` — startup/diagnostic messages via `log.Printf` /
  `log.Println`, fatal errors via `log.Fatalf`.
- `internal/api/middleware/logging.go` — request logger takes a `*log.Logger`
  and writes one preformatted line per request.
- Scattered `log.Printf` in `internal/api/handlers/auth.go`,
  `internal/service/user.go`, `internal/service/cleanup.go`.

There is no level filtering, no structured output, and no request correlation
— even though `chimiddleware.RequestID` is installed and `config.LogLevel` is
already parsed and validated (`debug|info|warn|error`) but consumed by nothing.

## Goals

- Adopt `log/slog` (standard library) as the single logging mechanism.
- Support the four levels `debug|info|warn|error` with runtime filtering driven
  by `config.LogLevel`.
- Structured output, selectable as human-readable text (default) or JSON, via a
  new `config.LogFormat`.
- Correlate every log line emitted during a request with its request ID (and
  user ID once auth resolves), automatically.
- Assign a deliberate level to every existing log call site.

## Non-goals (YAGNI)

- No `AddSource` (file:line) attribute.
- No log sampling, rotation, or file output.
- No per-component injected loggers — the global default logger is used
  (`slog.SetDefault`).
- No OpenTelemetry bridge.
- No change to `config.go`'s CLI-help output (stays on stderr) or to the
  pre-logger config-error path in `main` (`fmt.Fprintf(os.Stderr, …)` before the
  logger exists).

## Design

### 1. New package: `internal/logging`

Framework-agnostic, depends only on the standard library.

```go
package logging

// Options configures the process logger. Zero values are valid.
type Options struct {
	Level  string    // "debug"|"info"|"warn"|"error"; "" means "info"
	Format string    // "text"|"json"; "" means "text"
	Writer io.Writer // nil means os.Stdout
}

// New builds a *slog.Logger whose handler:
//   - filters at the configured level,
//   - formats as text or JSON,
//   - adds any attributes accumulated on the record's context (see ContextWith).
// It returns an error for an unrecognized Level or Format.
func New(opts Options) (*slog.Logger, error)
```

**`ContextHandler`** — a `slog.Handler` that wraps a base handler
(`slog.TextHandler` or `slog.JSONHandler`):

```go
type ContextHandler struct{ slog.Handler }

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if attrs := attrsFromContext(ctx); len(attrs) > 0 {
		r.AddAttrs(attrs...)
	}
	return h.Handler.Handle(ctx, r)
}

func (h ContextHandler) WithAttrs(a []slog.Attr) slog.Handler {
	return ContextHandler{h.Handler.WithAttrs(a)}
}

func (h ContextHandler) WithGroup(name string) slog.Handler {
	return ContextHandler{h.Handler.WithGroup(name)}
}
```

`WithAttrs` / `WithGroup` are overridden so the wrapper survives
`logger.With(...)`.

**Context bag** — attributes ride on the context and are flushed onto each
record by `ContextHandler`:

```go
type ctxKey struct{}

// ContextWith returns a context carrying attrs that ContextHandler adds to
// every record subsequently logged with a derived context. Repeated calls
// accumulate; later attrs with the same key shadow earlier ones per slog's
// normal last-wins rule.
func ContextWith(ctx context.Context, attrs ...slog.Attr) context.Context {
	prev, _ := ctx.Value(ctxKey{}).([]slog.Attr)
	next := make([]slog.Attr, 0, len(prev)+len(attrs))
	next = append(next, prev...)
	next = append(next, attrs...)
	return context.WithValue(ctx, ctxKey{}, next)
}

func attrsFromContext(ctx context.Context) []slog.Attr {
	a, _ := ctx.Value(ctxKey{}).([]slog.Attr)
	return a
}
```

`New` validates its inputs defensively even though `config` also validates —
the package must stand alone and be usable from tests.

### 2. `internal/config`

Add one field and one flag, mirroring the existing `LogLevel` code exactly.

- `Config.LogFormat string`
- Flag `log-format`, env `LOG_FORMAT`, default `"text"`, help text
  `"Log format: text|json (env: LOG_FORMAT)"`.
- In `validate()`: reject any value other than `text` or `json` with
  `config: log-format %q is not one of text|json`.

### 3. `cmd/server/main.go`

- Immediately after `config.Load` succeeds:

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

- Add a fatal helper and use it for every current `log.Fatalf`:

  ```go
  func fatal(msg string, args ...any) {
      slog.Error(msg, args...)
      os.Exit(1)
  }
  ```

  `setupRepositories` currently calls `log.Fatalf` directly and returns
  `nil, nil, nil` afterward to satisfy the compiler; keep that shape with
  `fatal(...)` in place of `log.Fatalf(...)`.

- Replace `log.Printf` / `log.Println` with `slog.Info` / `slog.Warn` per the
  level table below, using structured attributes where they add value
  (`version`, `addr`, `issuer`).

- `newHTTPServer`: set
  `ErrorLog: slog.NewLogLogger(slog.Default().Handler(), slog.LevelError)` so
  the stdlib server's internal errors are routed through slog.

- `newRouter`: delete the `log.New(os.Stdout, "[http] ", log.LstdFlags)`
  logger; call `middleware.RequestLogger(skipFn)` with no logger argument; add
  `middleware.RequestLogContext` on the outer router immediately after
  `chimiddleware.RequestID`.

- Remove the `"log"` import.

- `log.Fatal(s.ListenAndServe())` becomes:

  ```go
  slog.Info("starting inundated", "version", Version, "addr", addrStr)
  if err := s.ListenAndServe(); err != nil {
      fatal("http server stopped", "error", err)
  }
  ```

### 4. `internal/api/middleware`

**New `RequestLogContext`** (in a new `logcontext.go` or appended to
`logging.go`):

```go
func RequestLogContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id := chimiddleware.GetReqID(r.Context()); id != "" {
			r = r.WithContext(
				logging.ContextWith(r.Context(), slog.String("request_id", id)),
			)
		}
		next.ServeHTTP(w, r)
	})
}
```

Placed right after `chimiddleware.RequestID` so it — and therefore every
downstream handler, including `RequestLogger` — sees the seeded context.

**`RequestLogger`** signature changes to:

```go
func RequestLogger(skipFn func(r *http.Request) bool) func(http.Handler) http.Handler
```

Body: unchanged control flow, but the final line becomes

```go
level := slog.LevelInfo
if ww.Status() >= 500 {
	level = slog.LevelError
}
slog.LogAttrs(r.Context(), level, "http request",
	slog.String("method", r.Method),
	slog.String("path", r.URL.Path),
	slog.Int("status", ww.Status()),
	slog.Int("bytes", ww.BytesWritten()),
	slog.Duration("duration", time.Since(start)),
)
```

`request_id` is added by `ContextHandler` from the context seeded by
`RequestLogContext`; it is not passed explicitly.

**`OIDCAuth`** — after the user is resolved and before
`next.ServeHTTP(w, r.WithContext(ctx))`:

```go
ctx = logging.ContextWith(ctx, slog.String("user_id", user.Id.String()))
```

so logs emitted from authenticated handlers and the services they call carry
`user_id`. The request-summary line itself does not: `RequestLogger` sits
above `OIDCAuth` in the chain and its context predates this addition. This is
acceptable and intentional.

Import note: `internal/api/middleware` gains a dependency on
`internal/logging` (stdlib-only, no cycle).

### 5. Leaf call sites

| File / line | Before | After |
|---|---|---|
| `handlers/auth.go:108` | `log.Printf("OIDC callback failed: %v", err)` | `slog.WarnContext(ctx, "OIDC callback failed", "error", err)` |
| `service/user.go:79` | `log.Printf("first user %s adopted %d orphaned resources …", …)` | `slog.InfoContext(ctx, "first user adopted orphaned resources", "user_id", created.Id, "projects", adoption.Projects, "tags", adoption.Tags, "timespans", adoption.Timespans)` |
| `service/cleanup.go:50` | `log.Printf("cleanup: deleting expired sessions: %v", err)` | `slog.ErrorContext(ctx, "cleanup: deleting expired sessions failed", "error", err)` |
| `service/cleanup.go:53` | `log.Printf("cleanup: deleting expired login states: %v", err)` | `slog.ErrorContext(ctx, "cleanup: deleting expired login states failed", "error", err)` |
| `service/cleanup.go` (end of `cleanup`) | — (new) | `slog.DebugContext(ctx, "cleanup run complete")` |

Remove the now-unused `"log"` import from each of these files.

### Level assignment — `cmd/server/main.go`

| Message | Level |
|---|---|
| "Using in-memory repository (not recommended for production)" | Warn |
| "Using postgres repository" / "Applying database migrations…" | Info |
| "user registration disabled; only OIDC identities …" | Info |
| "registration is disabled and no users exist yet …" | Warn |
| "OIDC authentication enabled" (with `issuer` attr) | Info |
| "OIDC not configured; running in userless mode" | Info |
| "insecure cookies (Secure attribute off) …" | Warn |
| "TRUSTED_PROXIES not set …" | Warn |
| "starting inundated" (with `version`, `addr` attrs) | Info |
| migration failure / PG connect failure / unsupported DB URL / auth config error / `ListenAndServe` returned | Error, then `os.Exit(1)` |

## Testing

### `internal/logging/logging_test.go` (new)

- `New` with `Level: "warn"` drops an `Info` record, keeps a `Warn` record.
- `New` with `Format: "json"` emits parseable JSON; default / `"text"` emits
  `key=value` text.
- `New` with an unknown `Level` or `Format` returns an error.
- A logger from `New` writing with a context built by `ContextWith` includes
  those attributes in the output.
- `ContextWith` called twice accumulates both sets of attributes.
- `logger.With("k","v")` still routes through `ContextHandler` (context attrs
  still appear).

### `internal/api/middleware/logging_test.go` (new)

Capture output by calling `slog.SetDefault` with a `JSONHandler` over a
`bytes.Buffer` for the duration of each test (restore afterward).

- A normal request emits one `"http request"` record with `method`, `path`,
  `status`, `bytes` attributes.
- A handler returning 500 makes the record's level `ERROR`; a 200 is `INFO`.
- `skipFn` returning true emits nothing.
- With `RequestLogContext` in front and a request ID present, the record
  carries `request_id`.

### `internal/config/config_test.go`

- Default `LogFormat` is `"text"`.
- `LOG_FORMAT=json` (env) and `-log-format=json` (flag) are honored.
- `-log-format=yaml` is rejected by `validate()`.

### Regression

`cmd/server/main_test.go` compiles and passes unchanged (it exercises
`newRouter`, whose external behavior is unchanged). Full `go test ./...` and
`golangci-lint run` are green.

## Rollout

Single change set. New env var `LOG_FORMAT` is optional and defaults to the
current-equivalent human-readable output, so existing deployments are
unaffected. Document `LOG_LEVEL` (now actually wired) and `LOG_FORMAT` in
`README.md`'s configuration section.
