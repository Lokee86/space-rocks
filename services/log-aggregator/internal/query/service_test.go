package query

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

type captureEventStore struct {
	query  storage.Query
	result storage.QueryResult
	err    error
	calls  int
}

func (store *captureEventStore) AppendBatch(context.Context, []storage.Record) error {
	return nil
}

func (store *captureEventStore) Query(_ context.Context, query storage.Query) (storage.QueryResult, error) {
	store.calls++
	store.query = query
	return store.result, store.err
}

func (store *captureEventStore) Status(context.Context) (storage.Status, error) {
	return storage.Status{}, nil
}

func (store *captureEventStore) Close() error {
	return nil
}

func TestServiceMapsEveryFilterAndPreservesResult(t *testing.T) {
	auditRequired := true
	from := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	filter := Filter{
		From:               from,
		To:                 to,
		SchemaVersion:      "schema-1",
		Environment:        "test",
		BuildVersion:       "build-1",
		Service:            "api",
		ServiceInstanceID:  "instance-1",
		Category:           "http",
		Level:              "warn",
		Event:              "failed",
		TraceID:            "trace-1",
		RequestID:          "request-1",
		SessionID:          "session-1",
		RoomID:             "room-1",
		MatchID:            "match-1",
		PlayerID:           "player-1",
		AccountID:          "account-1",
		EventID:            "event-1",
		DiagnosticReportID: "report-1",
		AuditEventID:       "audit-1",
		IdempotencyKey:     "idempotency-1",
		AuditRequired:      &auditRequired,
		Limit:              7,
	}
	payload := json.RawMessage(`{"event_id":"event-1","nested":{"value":1}}`)
	store := &captureEventStore{
		result: storage.QueryResult{
			Records: []storage.Record{{Payload: payload}},
			Total:   3,
			Limited: true,
		},
	}

	result, err := NewService(store).Query(context.Background(), filter)
	if err != nil {
		t.Fatalf("query error: %v", err)
	}

	expected := storage.Query{
		From:               from,
		To:                 to,
		SchemaVersion:      "schema-1",
		Environment:        "test",
		BuildVersion:       "build-1",
		Service:            "api",
		ServiceInstanceID:  "instance-1",
		Category:           "http",
		Level:              "warn",
		Event:              "failed",
		TraceID:            "trace-1",
		RequestID:          "request-1",
		SessionID:          "session-1",
		RoomID:             "room-1",
		MatchID:            "match-1",
		PlayerID:           "player-1",
		AccountID:          "account-1",
		EventID:            "event-1",
		DiagnosticReportID: "report-1",
		AuditEventID:       "audit-1",
		IdempotencyKey:     "idempotency-1",
		AuditRequired:      &auditRequired,
		Limit:              7,
	}
	if store.calls != 1 {
		t.Fatalf("store calls=%d, want 1", store.calls)
	}
	if !reflect.DeepEqual(store.query, expected) {
		t.Fatalf("query=%+v, want %+v", store.query, expected)
	}
	if result.Total != 3 || !result.Limited || len(result.Events) != 1 || string(result.Events[0]) != string(payload) {
		t.Fatalf("result=%+v", result)
	}

	result.Events[0][0] = 'x'
	if string(payload) != `{"event_id":"event-1","nested":{"value":1}}` {
		t.Fatal("service did not deep-copy payload")
	}
}

func TestServicePreservesStorageOrderAndMetadata(t *testing.T) {
	first := json.RawMessage(`{"order":2}`)
	second := json.RawMessage(`{"order":1}`)
	store := &captureEventStore{
		result: storage.QueryResult{
			Records: []storage.Record{
				{Payload: first},
				{Payload: second},
			},
			Total:   4,
			Limited: true,
		},
	}

	result, err := NewService(store).Query(context.Background(), Filter{Limit: 2})
	if err != nil {
		t.Fatalf("query error: %v", err)
	}
	if store.calls != 1 || len(result.Events) != 2 || string(result.Events[0]) != string(first) || string(result.Events[1]) != string(second) || result.Total != 4 || !result.Limited {
		t.Fatalf("calls=%d result=%+v", store.calls, result)
	}
}

func TestServicePropagatesStorageError(t *testing.T) {
	want := errors.New("storage failed")
	store := &captureEventStore{err: want}

	_, err := NewService(store).Query(context.Background(), Filter{})
	if !errors.Is(err, want) {
		t.Fatalf("error=%v, want %v", err, want)
	}
	if store.calls != 1 {
		t.Fatalf("store calls=%d, want 1", store.calls)
	}
}
