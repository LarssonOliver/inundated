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
	Writer io.Writer // nil means os.Stdout
}

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
