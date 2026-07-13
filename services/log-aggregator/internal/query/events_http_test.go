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
}

func (f *fakeEventQuerier) Query(_ context.Context, filter Filter) (Result, error) {
	f.filter = filter
	return f.result, f.err
}

func TestEventsHandlerForwardsCompleteFilterAndRawEvents(t *testing.T) {
	fake := &fakeEventQuerier{result: Result{Events: []json.RawMessage{json.RawMessage(`{"event_id":"e1","message":"hello"}`)}, Total: 4, Limited: true}}
	req := httptest.NewRequest(http.MethodGet, EventsPath+"?from=2026-01-01T00:00:00Z&to=2026-01-02T00:00:00Z&service=api&service_instance_id=i1&level=warn&event=failed&trace_id=t&request_id=r&session_id=s&room_id=room&match_id=m&player_id=p&account_id=a&event_id=e&diagnostic_report_id=d&audit_event_id=ae&idempotency_key=k&audit_required=true&limit=20", nil)
	rec := httptest.NewRecorder()
	NewEventsHandler(fake).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || fake.filter.Service != "api" || fake.filter.ServiceInstanceID != "i1" || fake.filter.Level != "warn" || fake.filter.Event != "failed" || fake.filter.TraceID != "t" || fake.filter.RequestID != "r" || fake.filter.SessionID != "s" || fake.filter.RoomID != "room" || fake.filter.MatchID != "m" || fake.filter.PlayerID != "p" || fake.filter.AccountID != "a" || fake.filter.EventID != "e" || fake.filter.DiagnosticReportID != "d" || fake.filter.AuditEventID != "ae" || fake.filter.IdempotencyKey != "k" || fake.filter.Limit != 20 || fake.filter.AuditRequired == nil || !*fake.filter.AuditRequired {
		t.Fatalf("status=%d filter=%+v body=%s", rec.Code, fake.filter, rec.Body.String())
	}
	if !fake.filter.From.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) || !fake.filter.To.Equal(time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("filter=%+v", fake.filter)
	}
	var response EventsResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil || response.Total != 4 || !response.Limited || len(response.Events) != 1 || string(response.Events[0]) != `{"event_id":"e1","message":"hello"}` {
		t.Fatalf("response=%s err=%v", rec.Body.String(), err)
	}
}

func TestEventsHandlerRejectsUnknownMalformedAndInvertedFilters(t *testing.T) {
	for _, query := range []string{"?servcie=api", "?service=", "?service=api&service=worker", "?from=not-time", "?audit_required=maybe", "?limit=0", "?from=2026-01-02T00:00:00Z&to=2026-01-01T00:00:00Z"} {
		req := httptest.NewRequest(http.MethodGet, EventsPath+query, nil)
		rec := httptest.NewRecorder()
		NewEventsHandler(&fakeEventQuerier{}).ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "invalid_filter") {
			t.Fatalf("query=%s status=%d body=%s", query, rec.Code, rec.Body.String())
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

