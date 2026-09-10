package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Options struct {
	Level  string    // "debug"|"info"|"warn"|"error"; "" means "info"
	Format string    // "text"|"json"; "" means "text"
	Writer io.Writer // nil means os.Stdout, fronted by a non-blocking AsyncWriter
}

// New builds the application logger. When Writer is nil it logs to os.Stdout
// through an [AsyncWriter], so a stalled log destination can never block the
// goroutines that log (every request handler among them). The returned function
// flushes and stops that background writer; it is a no-op when Writer is set.
func New(opts Options) (*slog.Logger, func(), error) {
	level, err := parseLevel(opts.Level)
	if err != nil {
		return nil, nil, err
	}

	w := opts.Writer
	cleanup := func() {}
	if w == nil {
		async := NewAsyncWriter(os.Stdout)
		w = async
		cleanup = func() { _ = async.Close() }
	}

	handlerOpts := &slog.HandlerOptions{Level: level}

	var base slog.Handler
	switch strings.ToLower(opts.Format) {
	case "", "text":
		base = slog.NewTextHandler(w, handlerOpts)
	case "json":
		base = slog.NewJSONHandler(w, handlerOpts)
	default:
		cleanup()
		return nil, nil, fmt.Errorf("logging: unknown format %q (want text or json)", opts.Format)
	}

	return slog.New(&ContextHandler{Handler: base}), cleanup, nil
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
