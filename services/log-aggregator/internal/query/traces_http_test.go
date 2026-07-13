package query

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTraceHandlerForwardsUUIDLimitAndRawEvents(t *testing.T) {
	fake := &fakeEventQuerier{result: Result{Events: []json.RawMessage{json.RawMessage(`{"trace_id":"123e4567-e89b-12d3-a456-426614174000","message":"hello"}`)}, Total: 1}}
	req := httptest.NewRequest(http.MethodGet, TracesPathPrefix+"123e4567-e89b-12d3-a456-426614174000?limit=5", nil)
	rec := httptest.NewRecorder()
	NewTraceHandler(fake).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || fake.filter.TraceID != "123e4567-e89b-12d3-a456-426614174000" || fake.filter.Limit != 5 {
		t.Fatalf("status=%d filter=%+v body=%s", rec.Code, fake.filter, rec.Body.String())
	}
	var response EventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil || len(response.Events) != 1 || string(response.Events[0]) != `{"trace_id":"123e4567-e89b-12d3-a456-426614174000","message":"hello"}` {
		t.Fatalf("response=%s err=%v", rec.Body.String(), err)
	}
}

func TestTraceHandlerRejectsInvalidIDsPathsAndFilters(t *testing.T) {
	for _, path := range []string{
		"/v1/traces/123e4567-e89b-12d3-a456-426614174000/extra",
		"/v1/traces/123e4567-e89b-12d3-a456-426614174000/",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		NewTraceHandler(&fakeEventQuerier{}).ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("path=%s status=%d", path, rec.Code)
		}
	}
	for _, traceID := range []string{"not-a-uuid", "123e4567-e89b-12d3-a456-42661417400"} {
		req := httptest.NewRequest(http.MethodGet, TracesPathPrefix+traceID, nil)
		rec := httptest.NewRecorder()
		NewTraceHandler(&fakeEventQuerier{}).ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("trace_id=%s status=%d", traceID, rec.Code)
		}
	}
	for _, query := range []string{"?unknown=x", "?limit=1&limit=2", "?limit=0", "?limit=nope"} {
		req := httptest.NewRequest(http.MethodGet, TracesPathPrefix+"123e4567-e89b-12d3-a456-426614174000"+query, nil)
		rec := httptest.NewRecorder()
		NewTraceHandler(&fakeEventQuerier{}).ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("query=%s status=%d body=%s", query, rec.Code, rec.Body.String())
		}
	}
}

func TestTraceHandlerMethodRejection(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, TracesPathPrefix+"123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()
	NewTraceHandler(&fakeEventQuerier{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("status=%d allow=%q", rec.Code, rec.Header().Get("Allow"))
	}
}

func TestTraceHandlerServiceFailureIsSanitized(t *testing.T) {
	fake := &fakeEventQuerier{err: errors.New("secret storage details")}
	req := httptest.NewRequest(http.MethodGet, TracesPathPrefix+"123e4567-e89b-12d3-a456-426614174000", nil)
	rec := httptest.NewRecorder()
	NewTraceHandler(fake).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
