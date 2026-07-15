package httpapi

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	TraceIDHeader   = "X-Trace-ID"
	RequestIDHeader = "X-Request-ID"
)

type requestContextKey string

const (
	traceIDContextKey   requestContextKey = "player-data-trace-id"
	requestIDContextKey requestContextKey = "player-data-request-id"
)

// TraceIDFromContext returns the trace that owns the complete HTTP operation.
func TraceIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(traceIDContextKey).(string)
	return value
}

// RequestIDFromContext returns the unique identifier for one HTTP request.
func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(requestIDContextKey).(string)
	return value
}

// WithRequestContext assigns one trace and one request ID to each request.
// Invalid inbound traces are replaced with a new UUID.
func WithRequestContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, withRequestContext(w, r))
	})
}

func withRequestContext(w http.ResponseWriter, r *http.Request) *http.Request {
	if traceID, requestID := TraceIDFromContext(r.Context()), RequestIDFromContext(r.Context()); traceID != "" && requestID != "" {
		w.Header().Set(RequestIDHeader, requestID)
		return r
	}
	traceID := r.Header.Get(TraceIDHeader)
	if uuid.Validate(traceID) != nil {
		traceID = uuid.NewString()
	}
	requestID := uuid.NewString()
	w.Header().Set(RequestIDHeader, requestID)
	ctx := context.WithValue(r.Context(), traceIDContextKey, traceID)
	ctx = context.WithValue(ctx, requestIDContextKey, requestID)
	return r.WithContext(ctx)
}
