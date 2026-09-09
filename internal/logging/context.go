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
//
// Attributes are added via Record.AddAttrs, so if a caller has opened a group
// on the logger (logger.WithGroup(...)), the injected attributes land inside
// that group rather than at the record root. The app does not use groups today.
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
