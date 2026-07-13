package storagefake

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

func TestNewCopiesSeedRecordsAndPayloads(t *testing.T) {
	seed := recordForTest("seed", time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC), "seed-payload")
	store := New(seed)
	seed.Payload[0] = 'X'

	snapshot := store.Snapshot()
	if got := string(snapshot[0].Payload); got != "seed-payload" {
		t.Fatalf("seed payload was not copied: got %q", got)
	}
}

func TestAppendBatchPreservesOrderAndCopiesPayloads(t *testing.T) {
	store := New()
	records := []storage.Record{
		recordForTest("first", time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC), "one"),
		recordForTest("second", time.Date(2026, 7, 13, 10, 0, 1, 0, time.UTC), "two"),
	}
	if err := store.AppendBatch(context.Background(), records); err != nil {
		t.Fatal(err)
	}
	records[0].Payload[0] = 'X'

	snapshot := store.Snapshot()
	if got := snapshot[0].EventID; got != "first" {
		t.Fatalf("append order changed: got first event %q", got)
	}
	if got := snapshot[1].EventID; got != "second" {
		t.Fatalf("append order changed: got second event %q", got)
	}
	if got := string(snapshot[0].Payload); got != "one" {
		t.Fatalf("input payload was not copied: got %q", got)
	}
}

func TestSnapshotReturnsIndependentDeepCopy(t *testing.T) {
	store := New(recordForTest("one", time.Time{}, "payload"))
	snapshot := store.Snapshot()
	snapshot[0].Payload[0] = 'X'
	snapshot[0].EventID = "changed"

	fresh := store.Snapshot()
	if got := fresh[0].EventID; got != "one" {
		t.Fatalf("snapshot changed stored record: got %q", got)
	}
	if got := string(fresh[0].Payload); got != "payload" {
		t.Fatalf("snapshot changed stored payload: got %q", got)
	}
}

func TestStatusReportsReadinessCountsBytesAndBounds(t *testing.T) {
	oldest := time.Date(2026, 7, 13, 10, 0, 0, 0, time.UTC)
	newest := oldest.Add(2 * time.Second)
	store := New(
		recordForTest("old", oldest, "123"),
		recordForTest("new", newest, "12345"),
	)

	status, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ready || status.RecordCount != 2 || status.ByteCount != 8 {
		t.Fatalf("unexpected status: %+v", status)
	}
	if !status.Oldest.Equal(oldest) || !status.Newest.Equal(newest) {
		t.Fatalf("unexpected time bounds: oldest=%v newest=%v", status.Oldest, status.Newest)
	}
}

func TestCloseIsIdempotentAndRejectsOperations(t *testing.T) {
	store := New(recordForTest("one", time.Time{}, "payload"))
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	status, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Ready || status.RecordCount != 1 || status.ByteCount != 7 {
		t.Fatalf("unexpected closed status: %+v", status)
	}
	if err := store.AppendBatch(context.Background(), nil); !errors.Is(err, storage.ErrClosed) {
		t.Fatalf("AppendBatch error = %v, want %v", err, storage.ErrClosed)
	}
	if _, err := store.Query(context.Background(), storage.Query{}); !errors.Is(err, storage.ErrClosed) {
		t.Fatalf("Query error = %v, want %v", err, storage.ErrClosed)
	}
}

func TestCanceledContextsAndInjectedFailures(t *testing.T) {
	store := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.AppendBatch(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled AppendBatch error = %v", err)
	}
	if _, err := store.Query(ctx, storage.Query{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Query error = %v", err)
	}
	if _, err := store.Status(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled Status error = %v", err)
	}

	appendErr := errors.New("append failure")
	queryErr := errors.New("query failure")
	statusErr := errors.New("status failure")
	store.SetFailures(Failures{Append: appendErr, Query: queryErr, Status: statusErr})
	if err := store.AppendBatch(context.Background(), nil); !errors.Is(err, appendErr) {
		t.Fatalf("injected AppendBatch error = %v", err)
	}
	if _, err := store.Query(context.Background(), storage.Query{}); !errors.Is(err, queryErr) {
		t.Fatalf("injected Query error = %v", err)
	}
	if _, err := store.Status(context.Background()); !errors.Is(err, statusErr) {
		t.Fatalf("injected Status error = %v", err)
	}
}

// recordForTest is intentionally package-visible so other storagefake tests
// can share the same compact normalized record setup.
func recordForTest(eventID string, timestamp time.Time, payload string) storage.Record {
	return storage.Record{
		EventID:   eventID,
		Timestamp: timestamp,
		Event:     "test_event",
		Service:   "test-service",
		Payload:   []byte(payload),
	}
}
