package servicelog

import (
	"context"
	"log/slog"
)

// fanoutHandler forwards records and handler configuration to each child
// handler. It is intentionally concrete so output composition stays local to
// the servicelog runtime.
type fanoutHandler struct {
	handlers []slog.Handler
}

func newFanoutHandler(handlers ...slog.Handler) slog.Handler {
	return &fanoutHandler{handlers: append([]slog.Handler(nil), handlers...)}
}

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	var firstErr error
	for _, handler := range h.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		if err := handler.Handle(ctx, record); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	children := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		children[i] = handler.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: children}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	children := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		children[i] = handler.WithGroup(name)
	}
	return &fanoutHandler{handlers: children}
}
