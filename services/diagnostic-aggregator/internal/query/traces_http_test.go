package query

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testTraceID = "123e4567-e89b-12d3-a456-426614174000"

func TestTraceHandlerForwardsLimitAndRawEvents(t *testing.T) {
	fake := &fakeEventQuerier{result: Result{
		Events: []json.RawMessage{json.RawMessage(`{"trace_id":"123e4567-e89b-12d3-a456-426614174000","message":"hello"}`)},
		Total:  1,
	}}
	rec := httptest.NewRecorder()
	NewTraceHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, TracesPathPrefix+testTraceID+"?limit=5", nil))
	if rec.Code != http.StatusOK || fake.calls != 1 || fake.filter.TraceID != testTraceID || fake.filter.Limit != 5 {
		t.Fatalf("status=%d calls=%d filter=%+v", rec.Code, fake.calls, fake.filter)
	}
	var response EventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil || len(response.Events) != 1 || string(response.Events[0]) != `{"trace_id":"123e4567-e89b-12d3-a456-426614174000","message":"hello"}` {
		t.Fatalf("response=%s", rec.Body.String())
	}
}

func TestTraceHandlerDefaultAndConfiguredLimits(t *testing.T) {
	fake := &fakeEventQuerier{}
	NewTraceHandler(fake).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, TracesPathPrefix+testTraceID, nil))
	if fake.filter.Limit != 100 {
		t.Fatalf("default limit=%d", fake.filter.Limit)
	}
	fake = &fakeEventQuerier{}
	rec := httptest.NewRecorder()
	NewTraceHandlerWithPolicy(fake, LimitPolicy{Default: 2, Maximum: 3}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, TracesPathPrefix+testTraceID, nil))
	if rec.Code != http.StatusOK || fake.filter.Limit != 2 {
		t.Fatalf("status=%d limit=%d", rec.Code, fake.filter.Limit)
	}
	fake = &fakeEventQuerier{}
	rec = httptest.NewRecorder()
	NewTraceHandlerWithPolicy(fake, LimitPolicy{Default: 2, Maximum: 3}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, TracesPathPrefix+testTraceID+"?limit=3", nil))
	if rec.Code != http.StatusOK || fake.filter.Limit != 3 {
		t.Fatalf("status=%d limit=%d", rec.Code, fake.filter.Limit)
	}
}

func TestTraceHandlerRejectsInvalidRequestsWithoutQuerying(t *testing.T) {
	for _, query := range []string{"?limit=4", "?limit=1&limit=2", "?limit=nope", "?limit=0", "?limit=-1", "?unknown=x"} {
		fake := &fakeEventQuerier{}
		rec := httptest.NewRecorder()
		NewTraceHandlerWithPolicy(fake, LimitPolicy{Default: 2, Maximum: 3}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, TracesPathPrefix+testTraceID+query, nil))
		if rec.Code != http.StatusBadRequest || fake.calls != 0 {
			t.Fatalf("query=%s status=%d calls=%d", query, rec.Code, fake.calls)
		}
	}
	for _, id := range []string{"not-a-uuid", "123E4567-e89b-12d3-a456-426614174000"} {
		fake := &fakeEventQuerier{}
		rec := httptest.NewRecorder()
		NewTraceHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, TracesPathPrefix+id, nil))
		if rec.Code != http.StatusBadRequest || fake.calls != 0 {
			t.Fatalf("id=%s status=%d calls=%d", id, rec.Code, fake.calls)
		}
	}
}

func TestTraceHandlerRejectsInvalidPathsWithoutQuerying(t *testing.T) {
	for _, path := range []string{TracesPathPrefix + testTraceID + "/extra", TracesPathPrefix + testTraceID + "/"} {
		fake := &fakeEventQuerier{}
		rec := httptest.NewRecorder()
		NewTraceHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound || fake.calls != 0 {
			t.Fatalf("path=%s status=%d calls=%d", path, rec.Code, fake.calls)
		}
	}
}

func TestTraceHandlerMethodRejection(t *testing.T) {
	for _, traceID := range []string{testTraceID, "not-a-uuid"} {
		fake := &fakeEventQuerier{}
		rec := httptest.NewRecorder()
		NewTraceHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, TracesPathPrefix+traceID, nil))
		if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != http.MethodGet || fake.calls != 0 {
			t.Fatalf("trace_id=%s status=%d calls=%d", traceID, rec.Code, fake.calls)
		}
	}
}

func TestTraceHandlerServiceFailureIsSanitized(t *testing.T) {
	fake := &fakeEventQuerier{err: errors.New("secret storage details")}
	rec := httptest.NewRecorder()
	NewTraceHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, TracesPathPrefix+testTraceID, nil))
	if rec.Code != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
