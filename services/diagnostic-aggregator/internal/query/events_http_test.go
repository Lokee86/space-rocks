package query

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type fakeEventQuerier struct {
	filter Filter
	result Result
	err    error
	calls  int
}

func (f *fakeEventQuerier) Query(_ context.Context, filter Filter) (Result, error) {
	f.calls++
	f.filter = filter
	return f.result, f.err
}

func TestEventsHandlerForwardsCompleteFilterAndRawEvents(t *testing.T) {
	fake := &fakeEventQuerier{result: Result{
		Events: []json.RawMessage{json.RawMessage(`{"event_id":"e1","message":"hello"}`)},
		Total:   4,
		Limited: true,
	}}
	requestURL := EventsPath + "?from=2026-01-01T00:00:00Z&to=2026-01-02T00:00:00Z" +
		"&schema_version=schema-1&environment=test&build_version=build-1&category=http" +
		"&service=api&service_instance_id=123e4567-e89b-12d3-a456-426614174000" +
		"&level=warn&event=failed&trace_id=123e4567-e89b-12d3-a456-426614174001" +
		"&request_id=r&session_id=s&room_id=room&match_id=m&player_id=p&account_id=a" +
		"&event_id=123e4567-e89b-12d3-a456-426614174002" +
		"&diagnostic_report_id=123e4567-e89b-12d3-a456-426614174003" +
		"&audit_event_id=123e4567-e89b-12d3-a456-426614174004" +
		"&idempotency_key=k&audit_required=true&limit=20"
	rec := httptest.NewRecorder()
	NewEventsHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, requestURL, nil))

	if rec.Code != http.StatusOK || fake.calls != 1 {
		t.Fatalf("status=%d calls=%d body=%s", rec.Code, fake.calls, rec.Body.String())
	}
	if fake.filter.SchemaVersion != "schema-1" || fake.filter.Environment != "test" || fake.filter.BuildVersion != "build-1" || fake.filter.Category != "http" || fake.filter.Service != "api" || fake.filter.ServiceInstanceID != "123e4567-e89b-12d3-a456-426614174000" || fake.filter.Level != "warn" || fake.filter.Event != "failed" || fake.filter.TraceID != "123e4567-e89b-12d3-a456-426614174001" || fake.filter.RequestID != "r" || fake.filter.SessionID != "s" || fake.filter.RoomID != "room" || fake.filter.MatchID != "m" || fake.filter.PlayerID != "p" || fake.filter.AccountID != "a" || fake.filter.EventID != "123e4567-e89b-12d3-a456-426614174002" || fake.filter.DiagnosticReportID != "123e4567-e89b-12d3-a456-426614174003" || fake.filter.AuditEventID != "123e4567-e89b-12d3-a456-426614174004" || fake.filter.IdempotencyKey != "k" || fake.filter.Limit != 20 || fake.filter.AuditRequired == nil || !*fake.filter.AuditRequired {
		t.Fatalf("filter=%+v", fake.filter)
	}
	if !fake.filter.From.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) || !fake.filter.To.Equal(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("filter=%+v", fake.filter)
	}
	var response EventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil || response.Total != 4 || !response.Limited || len(response.Events) != 1 || string(response.Events[0]) != `{"event_id":"e1","message":"hello"}` {
		t.Fatalf("response=%s err=%v", rec.Body.String(), err)
	}
}

func TestEventsHandlerDefaultAndConfiguredLimits(t *testing.T) {
	defaultFake := &fakeEventQuerier{}
	NewEventsHandler(defaultFake).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, EventsPath, nil))
	if defaultFake.filter.Limit != 100 {
		t.Fatalf("default limit=%d", defaultFake.filter.Limit)
	}

	configuredFake := &fakeEventQuerier{}
	rec := httptest.NewRecorder()
	NewEventsHandlerWithPolicy(configuredFake, LimitPolicy{Default: 7, Maximum: 9}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, EventsPath, nil))
	if rec.Code != http.StatusOK || configuredFake.filter.Limit != 7 {
		t.Fatalf("status=%d limit=%d", rec.Code, configuredFake.filter.Limit)
	}
	configuredFake = &fakeEventQuerier{}
	rec = httptest.NewRecorder()
	NewEventsHandlerWithPolicy(configuredFake, LimitPolicy{Default: 7, Maximum: 9}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, EventsPath+"?limit=9", nil))
	if rec.Code != http.StatusOK || configuredFake.filter.Limit != 9 {
		t.Fatalf("status=%d limit=%d", rec.Code, configuredFake.filter.Limit)
	}
}

func TestEventsHandlerRejectsInvalidLimitsWithoutQuerying(t *testing.T) {
	for _, query := range []string{"?limit=0", "?limit=-1", "?limit=nope", "?limit=1&limit=2", "?limit=1001"} {
		fake := &fakeEventQuerier{}
		rec := httptest.NewRecorder()
		NewEventsHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, EventsPath+query, nil))
		if rec.Code != http.StatusBadRequest || fake.calls != 0 {
			t.Fatalf("query=%s status=%d calls=%d", query, rec.Code, fake.calls)
		}
	}
}

func TestEventsHandlerRejectsNonCanonicalUUIDFiltersWithoutQuerying(t *testing.T) {
	for _, field := range []string{"service_instance_id", "trace_id", "event_id", "diagnostic_report_id", "audit_event_id"} {
		fake := &fakeEventQuerier{}
		rec := httptest.NewRecorder()
		query := "?" + field + "=123E4567-e89b-12d3-a456-426614174000"
		NewEventsHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, EventsPath+query, nil))
		if rec.Code != http.StatusBadRequest || fake.calls != 0 {
			t.Fatalf("field=%s status=%d calls=%d", field, rec.Code, fake.calls)
		}
	}
}

func TestEventsHandlerRejectsUnknownMalformedAndInvertedFilters(t *testing.T) {
	for _, query := range []string{"?servcie=api", "?service=", "?service=api&service=worker", "?from=not-time", "?audit_required=maybe", "?from=2026-01-02T00:00:00Z&to=2026-01-01T00:00:00Z"} {
		fake := &fakeEventQuerier{}
		rec := httptest.NewRecorder()
		NewEventsHandler(fake).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, EventsPath+query, nil))
		if rec.Code != http.StatusBadRequest || fake.calls != 0 || !strings.Contains(rec.Body.String(), "invalid_filter") {
			t.Fatalf("query=%s status=%d calls=%d body=%s", query, rec.Code, fake.calls, rec.Body.String())
		}
	}
}

func TestEventsHandlerMethodRejection(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, EventsPath, nil)
	rec := httptest.NewRecorder()
	NewEventsHandler(&fakeEventQuerier{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("status=%d allow=%q", rec.Code, rec.Header().Get("Allow"))
	}
}

func TestEventsHandlerServiceFailureIsSanitized(t *testing.T) {
	fake := &fakeEventQuerier{err: errors.New("secret storage details")}
	req := httptest.NewRequest(http.MethodGet, EventsPath, nil)
	rec := httptest.NewRecorder()
	NewEventsHandler(fake).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
