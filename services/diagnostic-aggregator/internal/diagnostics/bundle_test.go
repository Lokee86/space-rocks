package diagnostics

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

const testTraceID = "123e4567-e89b-12d3-a456-426614174000"
const testReportID = "123e4567-e89b-12d3-a456-426614174001"

type fakeQuerier struct {
	result storage.QueryResult
	err    error
	query  storage.Query
}

func (f *fakeQuerier) Query(_ context.Context, query storage.Query) (storage.QueryResult, error) {
	f.query = query
	return f.result, f.err
}

func testBuilder(querier EventQuerier) Builder {
	return Builder{
		Store:   querier,
		NewUUID: func() (string, error) { return testReportID, nil },
		Now:     func() time.Time { return time.Unix(10, 0).In(time.FixedZone("test", 3600)) },
		Sanitize: func(payload json.RawMessage) (json.RawMessage, SanitizationSummary, error) {
			result := append(json.RawMessage(`{"safe":true,"payload":`), payload...)
			result = append(result, '}')
			return result, SanitizationSummary{RedactedFields: 1, DroppedFields: 2}, nil
		},
	}
}

func testRecord(id, service string, timestamp time.Time) storage.Record {
	return storage.Record{
		EventID:           id,
		Service:           service,
		ServiceInstanceID: "instance",
		Environment:       "test",
		BuildVersion:      "build",
		RequestID:         "request",
		SessionID:         "session",
		RoomID:            "room",
		MatchID:           "match",
		PlayerID:          "player",
		AccountID:         "account",
		Timestamp:         timestamp,
		Payload:           json.RawMessage(`{"secret":"x"}`),
	}
}

func TestBuilderBuildsRawJSONBundleAndSummaries(t *testing.T) {
	querier := &fakeQuerier{result: storage.QueryResult{
		Records: []storage.Record{
			testRecord("2", "z", time.Unix(2, 0)),
			testRecord("1", "a", time.Unix(1, 0)),
		},
		Total: 2,
	}}

	bundle, err := testBuilder(querier).Build(context.Background(), testTraceID, 10)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(bundle)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatal(err)
	}
	if _, ok := document["events"].([]any); !ok {
		t.Fatalf("events did not encode as an array: %s", encoded)
	}
	if bundle.Events[0].Payload[0] != '{' || bundle.Services[0] != "a" || bundle.EventTimeRange.From.Unix() != 1 {
		t.Fatalf("unexpected bundle: %#v", bundle)
	}
	if bundle.CreatedAt.Location() != time.UTC {
		t.Fatalf("created time is not UTC: %v", bundle.CreatedAt.Location())
	}
	if bundle.Sanitization.RedactedFields != 2 || bundle.Sanitization.DroppedFields != 4 {
		t.Fatalf("unexpected sanitization summary: %#v", bundle.Sanitization)
	}
	if querier.query.TraceID != testTraceID || querier.query.Limit != 10 {
		t.Fatalf("unexpected query: %#v", querier.query)
	}
}

func TestBuilderPreservesStorageOrderAndCopiesPayloads(t *testing.T) {
	payload := json.RawMessage(`{"value":"original"}`)
	querier := &fakeQuerier{result: storage.QueryResult{Records: []storage.Record{
		{EventID: "first", Timestamp: time.Unix(1, 0), Payload: payload},
		{EventID: "second", Timestamp: time.Unix(2, 0), Payload: json.RawMessage(`{"value":2}`)},
	}, Total: 2}}
	builder := testBuilder(querier)
	builder.Sanitize = func(input json.RawMessage) (json.RawMessage, SanitizationSummary, error) {
		return append(json.RawMessage(nil), input...), SanitizationSummary{}, nil
	}

	bundle, err := builder.Build(context.Background(), testTraceID, 2)
	if err != nil {
		t.Fatal(err)
	}
	payload[0] = 'x'
	if bundle.Events[0].EventID != "first" || bundle.Events[1].EventID != "second" || bundle.Events[0].Payload[0] != '{' {
		t.Fatalf("storage order or copy was lost: %#v", bundle.Events)
	}
}

func TestBuilderValidationAndConfiguration(t *testing.T) {
	querier := &fakeQuerier{}
	builder := testBuilder(querier)
	builder.Sanitize = nil
	if _, err := builder.Build(context.Background(), testTraceID, 1); !errors.Is(err, ErrSanitizerRequired) {
		t.Fatalf("expected sanitizer error, got %v", err)
	}
	if _, err := testBuilder(querier).Build(context.Background(), "bad", 1); !errors.Is(err, ErrInvalidTraceID) {
		t.Fatalf("expected trace id error, got %v", err)
	}
	if _, err := testBuilder(querier).Build(context.Background(), testTraceID, -1); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("expected limit error, got %v", err)
	}
	if _, err := testBuilder(querier).Build(context.Background(), testTraceID, 1); !errors.Is(err, ErrNoEvents) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestBuilderPropagatesFailuresAndRejectsGeneratedID(t *testing.T) {
	querier := &fakeQuerier{result: storage.QueryResult{Records: []storage.Record{{Payload: json.RawMessage(`{"x":1}`)}}, Total: 1}}
	builder := testBuilder(querier)
	builder.NewUUID = func() (string, error) { return "", errors.New("uuid failure") }
	if _, err := builder.Build(context.Background(), testTraceID, 1); err == nil {
		t.Fatal("uuid failure was lost")
	}
	builder.NewUUID = func() (string, error) { return "not-a-uuid", nil }
	if _, err := builder.Build(context.Background(), testTraceID, 1); !errors.Is(err, ErrInvalidReportID) {
		t.Fatalf("expected invalid report id, got %v", err)
	}
	builder.NewUUID = func() (string, error) { return testReportID, nil }
	builder.Sanitize = func(json.RawMessage) (json.RawMessage, SanitizationSummary, error) {
		return nil, SanitizationSummary{}, errors.New("sanitize failure")
	}
	if _, err := builder.Build(context.Background(), testTraceID, 1); err == nil {
		t.Fatal("sanitizer failure was lost")
	}
	querier.err = errors.New("query failure")
	if _, err := builder.Build(context.Background(), testTraceID, 1); err == nil {
		t.Fatal("query failure was lost")
	}
}

func TestBuilderLimitDefaultsAndMaximum(t *testing.T) {
	querier := &fakeQuerier{result: storage.QueryResult{Records: []storage.Record{{Payload: json.RawMessage(`{"x":1}`)}}, Total: 1}}
	builder := testBuilder(querier)
	builder.DefaultLimit = 2
	builder.MaximumLimit = 3
	if _, err := builder.Build(context.Background(), testTraceID, 0); err != nil || querier.query.Limit != 2 {
		t.Fatalf("default limit: %v, query: %#v", err, querier.query)
	}
	if _, err := builder.Build(context.Background(), testTraceID, 4); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("expected maximum limit error, got %v", err)
	}
}

func TestBuilderRejectsInvalidSanitizedPayload(t *testing.T) {
	for _, sanitized := range []json.RawMessage{
		nil,
		json.RawMessage(`{"invalid"`),
		json.RawMessage(`[]`),
		json.RawMessage(`"scalar"`),
		json.RawMessage(`null`),
	} {
		t.Run(string(sanitized), func(t *testing.T) {
			querier := &fakeQuerier{result: storage.QueryResult{Records: []storage.Record{{Payload: json.RawMessage(`{"x":1}`)}}, Total: 1}}
			builder := testBuilder(querier)
			builder.Sanitize = func(json.RawMessage) (json.RawMessage, SanitizationSummary, error) {
				return sanitized, SanitizationSummary{}, nil
			}
			if _, err := builder.Build(context.Background(), testTraceID, 1); !errors.Is(err, ErrInvalidSanitizedPayload) {
				t.Fatalf("expected invalid sanitized payload error, got %v", err)
			}
		})
	}
}
