package logging

import (
	"context"
	"log/slog"
)

type fanoutHandler struct {
	handlers []slog.Handler
}

func (handler fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, child := range handler.handlers {
		if child.Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (handler fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	var firstErr error

	for _, child := range handler.handlers {
		if !child.Enabled(ctx, record.Level) {
			continue
		}

		childRecord := record.Clone()
		if err := child.Handle(ctx, childRecord); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (handler fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, 0, len(handler.handlers))
	for _, child := range handler.handlers {
		next = append(next, child.WithAttrs(attrs))
	}

	return fanoutHandler{handlers: next}
}

func (handler fanoutHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, 0, len(handler.handlers))
	for _, child := range handler.handlers {
		next = append(next, child.WithGroup(name))
	}

	return fanoutHandler{handlers: next}
}
