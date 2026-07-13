package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/health"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/ingest"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/query"
)

type instrumentIngestor struct {
	result ingest.BatchResult
	err    error
}

func (i instrumentIngestor) IngestBatch(context.Context, []json.RawMessage) (ingest.BatchResult, error) {
	return i.result, i.err
}

type instrumentQuerier struct {
	result query.Result
	err    error
}

func (q instrumentQuerier) Query(context.Context, query.Filter) (query.Result, error) {
	return q.result, q.err
}

func TestInstrumentedIngestorCountsAndPreservesResult(t *testing.T) {
	state := health.NewState("i", "v", "e", time.Time{})
	want := ingest.BatchResult{BatchID: "b", Accepted: 3, Rejected: []ingest.EventRejection{{Code: "bad"}, {Code: "redaction_rejected"}, {Code: "redaction_rejected"}}}
	wrapped := NewInstrumentedIngestor(instrumentIngestor{result: want}, state)
	got, err := wrapped.IngestBatch(context.Background(), nil)
	if err != nil || got.BatchID != want.BatchID || got.Accepted != want.Accepted || len(got.Rejected) != len(want.Rejected) {
		t.Fatalf("result = %+v, err = %v", got, err)
	}
	snapshot := state.Snapshot()
	if snapshot.BatchesReceived != 1 || snapshot.EventsAccepted != 3 || snapshot.EventsRejected != 3 || snapshot.EventsRedacted != 2 {
		t.Fatalf("unexpected counters: %+v", snapshot)
	}
}

func TestInstrumentedIngestorPreservesErrorsWithoutCounting(t *testing.T) {
	state := health.NewState("i", "v", "e", time.Time{})
	want := errors.New("failed")
	_, err := NewInstrumentedIngestor(instrumentIngestor{err: want}, state).IngestBatch(context.Background(), nil)
	if !errors.Is(err, want) || state.Snapshot().BatchesReceived != 1 {
		t.Fatalf("err=%v snapshot=%+v", err, state.Snapshot())
	}
}

func TestInstrumentedQuerierCountsFailuresAndPreservesValues(t *testing.T) {
	state := health.NewState("i", "v", "e", time.Time{})
	want := errors.New("query failed")
	wrapped := NewInstrumentedQuerier(instrumentQuerier{err: want}, state)
	_, err := wrapped.Query(context.Background(), query.Filter{})
	if !errors.Is(err, want) || state.Snapshot().QueryFailures != 1 {
		t.Fatalf("err=%v snapshot=%+v", err, state.Snapshot())
	}
	result := query.Result{Total: 4}
	got, err := NewInstrumentedQuerier(instrumentQuerier{result: result}, state).Query(context.Background(), query.Filter{})
	if err != nil || got.Total != result.Total || state.Snapshot().QueryFailures != 1 {
		t.Fatalf("result=%+v err=%v snapshot=%+v", got, err, state.Snapshot())
	}
}
