package ingest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type fakeBatchIngestor struct {
	events []json.RawMessage
	result BatchResult
	err    error
}

func (f *fakeBatchIngestor) IngestBatch(_ context.Context, events []json.RawMessage) (BatchResult, error) {
	f.events = events
	return f.result, f.err
}

func TestHandlerSuccessForwardsRawEvents(t *testing.T) {
	fake := &fakeBatchIngestor{result: BatchResult{BatchID: "batch-1", Accepted: 1, Rejected: []EventRejection{{Index: 1, EventID: "e2", Code: "invalid"}}}}
	req := httptest.NewRequest(http.MethodPost, EventsPath, bytes.NewBufferString(`{"events":[{"event_id":"e1","message":"keep me"},{"event_id":"e2","n":2}]}`))
	rec := httptest.NewRecorder()
	NewHandler(fake).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK || len(fake.events) != 2 || string(fake.events[0]) != `{"event_id":"e1","message":"keep me"}` {
		t.Fatalf("status=%d events=%+v body=%s", rec.Code, fake.events, rec.Body.String())
	}
	var response BatchResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil || response.BatchID != "batch-1" || response.Accepted != 1 || response.Rejected != 1 || len(response.Rejections) != 1 || response.Rejections[0].Index != 1 || response.Rejections[0].EventID != "e2" || response.Rejections[0].Code != "invalid" {
		t.Fatalf("response=%s err=%v", rec.Body.String(), err)
	}
}

func TestHandlerRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"malformed", `{"events":`},
		{"unknown", `{"events":[],"extra":true}`},
		{"trailing", `{"events":[{"message":"keep me"}]} {}`},
		{"empty", `{"events":[]}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, EventsPath, bytes.NewBufferString(tc.body))
			rec := httptest.NewRecorder()
			NewHandler(&fakeBatchIngestor{}).ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandlerMethodRejection(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, EventsPath, nil)
	rec := httptest.NewRecorder()
	NewHandler(&fakeBatchIngestor{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed || rec.Header().Get("Allow") != http.MethodPost {
		t.Fatalf("status=%d allow=%q", rec.Code, rec.Header().Get("Allow"))
	}
}

func TestHandlerServiceFailureIsStructured(t *testing.T) {
	fake := &fakeBatchIngestor{err: errors.New("secret database details")}
	req := httptest.NewRequest(http.MethodPost, EventsPath, bytes.NewBufferString(`{"events":[{"message":"keep me"}]}`))
	rec := httptest.NewRecorder()
	NewHandler(fake).ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable || rec.Body.String() == "" || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
