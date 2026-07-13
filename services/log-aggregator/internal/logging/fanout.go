package logging

import (
	"context"
	"log/slog"
)

type fanout struct{ handlers []slog.Handler }

func newFanout(handlers ...slog.Handler) slog.Handler { return fanout{handlers: handlers} }

func (f fanout) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range f.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f fanout) Handle(ctx context.Context, record slog.Record) error {
	var first error
	for _, handler := range f.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		if err := handler.Handle(ctx, record); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (f fanout) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(f.handlers))
	for index, handler := range f.handlers {
		handlers[index] = handler.WithAttrs(attrs)
	}
	return fanout{handlers: handlers}
}

func (f fanout) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(f.handlers))
	for index, handler := range f.handlers {
		handlers[index] = handler.WithGroup(name)
	}
	return fanout{handlers: handlers}
}
