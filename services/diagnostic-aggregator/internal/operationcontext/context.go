package operationcontext

import "context"

// Values identifies one diagnostic-aggregator operation. Values are stored by
// value so callers cannot mutate context owned by another boundary.
type Values struct {
	TraceID   string
	RequestID string
	Route     string
}

type contextKey struct{}

func With(ctx context.Context, values Values) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, contextKey{}, values)
}

func From(ctx context.Context) (Values, bool) {
	if ctx == nil {
		return Values{}, false
	}
	values, ok := ctx.Value(contextKey{}).(Values)
	return values, ok
}
