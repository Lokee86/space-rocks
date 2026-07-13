package storagefake

import (
	"context"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

func TestQueryHonorsEveryStringFilter(t *testing.T) {
	record := recordForTest("event-1", time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC), "payload")
	record.SchemaVersion = "schema-1"
	record.Environment = "test"
	record.BuildVersion = "build-1"
	record.Category = "network"
	record.ServiceInstanceID = "instance-1"
	record.Level = "warn"
	record.Event = "connection_failed"
	record.TraceID = "trace-1"
	record.RequestID = "request-1"
	record.SessionID = "session-1"
	record.RoomID = "room-1"
	record.MatchID = "match-1"
	record.PlayerID = "Player-1"
	record.AccountID = "account-1"
	record.DiagnosticReportID = "report-1"
	record.AuditEventID = "audit-1"
	record.IdempotencyKey = "idempotency-1"
	store := New(record)

	query := storage.Query{
		SchemaVersion:      "schema-1",
		Environment:        "test",
		BuildVersion:       "build-1",
		Category:           "network",
		Service:            "test-service",
		ServiceInstanceID:  "instance-1",
		Level:              "warn",
		Event:              "connection_failed",
		TraceID:            "trace-1",
		RequestID:          "request-1",
		SessionID:          "session-1",
		RoomID:             "room-1",
		MatchID:            "match-1",
		PlayerID:           "Player-1",
		AccountID:          "account-1",
		EventID:            "event-1",
		DiagnosticReportID: "report-1",
		AuditEventID:       "audit-1",
		IdempotencyKey:     "idempotency-1",
	}
	result, err := store.Query(context.Background(), query)
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || len(result.Records) != 1 || result.Records[0].EventID != "event-1" {
		t.Fatalf("unexpected filtered result: %+v", result)
	}

	tests := []struct {
		name   string
		mutate func(*storage.Query)
	}{
		{"schema_version", func(query *storage.Query) { query.SchemaVersion = "other" }},
		{"environment", func(query *storage.Query) { query.Environment = "other" }},
		{"build_version", func(query *storage.Query) { query.BuildVersion = "other" }},
		{"category", func(query *storage.Query) { query.Category = "other" }},
		{"service", func(query *storage.Query) { query.Service = "other" }},
		{"service_instance_id", func(query *storage.Query) { query.ServiceInstanceID = "other" }},
		{"level", func(query *storage.Query) { query.Level = "other" }},
		{"event_name", func(query *storage.Query) { query.Event = "other" }},
		{"trace_id", func(query *storage.Query) { query.TraceID = "other" }},
		{"request_id", func(query *storage.Query) { query.RequestID = "other" }},
		{"session_id", func(query *storage.Query) { query.SessionID = "other" }},
		{"room_id", func(query *storage.Query) { query.RoomID = "other" }},
		{"match_id", func(query *storage.Query) { query.MatchID = "other" }},
		{"player_id", func(query *storage.Query) { query.PlayerID = "other" }},
		{"account_id", func(query *storage.Query) { query.AccountID = "other" }},
		{"event_id", func(query *storage.Query) { query.EventID = "other" }},
		{"diagnostic_report_id", func(query *storage.Query) { query.DiagnosticReportID = "other" }},
		{"audit_event_id", func(query *storage.Query) { query.AuditEventID = "other" }},
		{"idempotency_key", func(query *storage.Query) { query.IdempotencyKey = "other" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mismatched := query
			test.mutate(&mismatched)
			result, err := store.Query(context.Background(), mismatched)
			if err != nil {
				t.Fatal(err)
			}
			if result.Total != 0 || len(result.Records) != 0 {
				t.Fatalf("mismatched filter returned records: %+v", result)
			}
		})
	}
}

func TestQueryTimeBoundsAreInclusive(t *testing.T) {
	from := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)
	to := from.Add(2 * time.Second)
	store := New(
		recordForTest("before", from.Add(-time.Second), "before"),
		recordForTest("from", from, "from"),
		recordForTest("to", to, "to"),
		recordForTest("after", to.Add(time.Second), "after"),
	)

	result, err := store.Query(context.Background(), storage.Query{From: from, To: to})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 2 || len(result.Records) != 2 {
		t.Fatalf("unexpected bounded result: %+v", result)
	}
	if result.Records[0].EventID != "from" || result.Records[1].EventID != "to" {
		t.Fatalf("bounds were not inclusive: %+v", result.Records)
	}
}

func TestQueryAuditRequiredFiltersTrueAndFalse(t *testing.T) {
	trueValue := true
	falseValue := false
	trueStore := New(recordWithAudit("true", true), recordWithAudit("false", false))

	for name, filter := range map[string]*bool{"true": &trueValue, "false": &falseValue} {
		t.Run(name, func(t *testing.T) {
			result, err := trueStore.Query(context.Background(), storage.Query{AuditRequired: filter})
			if err != nil {
				t.Fatal(err)
			}
			if result.Total != 1 || result.Records[0].EventID != name {
				t.Fatalf("unexpected audit result: %+v", result)
			}
		})
	}
}

func TestQueryOrdersChronologicallyAndBreaksTiesByEventID(t *testing.T) {
	timestamp := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)
	store := New(
		recordForTest("z-last", timestamp, "z"),
		recordForTest("a-first", timestamp, "a"),
		recordForTest("middle", timestamp.Add(-time.Second), "m"),
	)

	result, err := store.Query(context.Background(), storage.Query{})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"middle", "a-first", "z-last"}
	for index, eventID := range want {
		if result.Records[index].EventID != eventID {
			t.Fatalf("record %d = %q, want %q", index, result.Records[index].EventID, eventID)
		}
	}
}

func TestQueryLimitReportsTotalBeforeLimit(t *testing.T) {
	store := New(
		recordForTest("one", time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC), "1"),
		recordForTest("two", time.Date(2026, 7, 13, 10, 0, 1, 0, time.UTC), "2"),
		recordForTest("three", time.Date(2026, 7, 13, 10, 0, 2, 0, time.UTC), "3"),
	)

	result, err := store.Query(context.Background(), storage.Query{Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 3 || len(result.Records) != 2 || !result.Limited {
		t.Fatalf("unexpected limited result: %+v", result)
	}
	if result.Records[0].EventID != "one" || result.Records[1].EventID != "two" {
		t.Fatalf("limit changed ordering: %+v", result.Records)
	}

	unlimited, err := store.Query(context.Background(), storage.Query{Limit: 0})
	if err != nil {
		t.Fatal(err)
	}
	if unlimited.Total != 3 || len(unlimited.Records) != 3 || unlimited.Limited {
		t.Fatalf("non-positive limit was applied: %+v", unlimited)
	}
}

func TestQueryReturnsPayloadDeepCopies(t *testing.T) {
	store := New(recordForTest("one", time.Time{}, "payload"))
	result, err := store.Query(context.Background(), storage.Query{})
	if err != nil {
		t.Fatal(err)
	}
	result.Records[0].Payload[0] = 'X'

	snapshot := store.Snapshot()
	if got := string(snapshot[0].Payload); got != "payload" {
		t.Fatalf("query payload mutated stored record: got %q", got)
	}
}

func recordWithAudit(eventID string, required bool) storage.Record {
	record := recordForTest(eventID, time.Time{}, eventID)
	record.AuditRequired = required
	return record
}
