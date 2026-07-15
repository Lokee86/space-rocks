package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestWithRequestContextCreatesAndReturnsIDs(t *testing.T) {
	var traceID, requestID string
	handler := WithRequestContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID = TraceIDFromContext(r.Context())
		requestID = RequestIDFromContext(r.Context())
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if uuid.Validate(traceID) != nil || uuid.Validate(requestID) != nil {
		t.Fatalf("expected UUID trace and request IDs, got %q and %q", traceID, requestID)
	}
	if got := recorder.Header().Get(RequestIDHeader); got != requestID {
		t.Fatalf("response request ID = %q, want %q", got, requestID)
	}
}

func TestWithRequestContextPreservesValidTraceAndReplacesInvalidTrace(t *testing.T) {
	validTrace := uuid.NewString()
	for _, tc := range []struct {
		name      string
		inbound   string
		wantTrace string
	}{
		{name: "valid", inbound: validTrace, wantTrace: validTrace},
		{name: "invalid", inbound: "not-a-uuid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var gotTrace string
			handler := WithRequestContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotTrace = TraceIDFromContext(r.Context())
			}))
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			request.Header.Set(TraceIDHeader, tc.inbound)
			handler.ServeHTTP(recorder, request)

			if tc.wantTrace != "" && gotTrace != tc.wantTrace {
				t.Fatalf("trace ID = %q, want %q", gotTrace, tc.wantTrace)
			}
			if uuid.Validate(gotTrace) != nil {
				t.Fatalf("trace ID %q is not a UUID", gotTrace)
			}
		})
	}
}

func TestWithRequestContextUsesOneRequestIDWhenAppliedTwice(t *testing.T) {
	var first, second string
	handler := WithRequestContext(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		first = RequestIDFromContext(r.Context())
		r = withRequestContext(w, r)
		second = RequestIDFromContext(r.Context())
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	if first == "" || first != second {
		t.Fatalf("request ID changed during one request: %q then %q", first, second)
	}
}
