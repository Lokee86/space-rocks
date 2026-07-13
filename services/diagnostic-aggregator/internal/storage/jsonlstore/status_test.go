package jsonlstore

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

func TestStatusIncludesBufferedActiveRecordsAndTimes(t *testing.T) {
	root := t.TempDir()
	store, err := NewWithClock(testConfig(root), fixedClock{now: time.Unix(100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	old, newer := time.Unix(10, 0), time.Unix(20, 0)
	records := []storage.Record{{EventID: "old", Timestamp: old}, {EventID: "new", Timestamp: newer}}
	if err := store.AppendBatch(context.Background(), records); err != nil {
		t.Fatal(err)
	}
	status, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ready || status.RecordCount != 2 || !status.Oldest.Equal(old) || !status.Newest.Equal(newer) {
		t.Fatalf("status = %+v", status)
	}
}

func TestStatusReportsPayloadBytesAndAcceptsNilContext(t *testing.T) {
	root := t.TempDir()
	store, err := NewWithClock(testConfig(root), fixedClock{now: time.Unix(100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	payload := []byte("{\"key\":\"value\"}")
	if !json.Valid(payload) {
		t.Fatal("valid payload fixture is not valid JSON")
	}
	if err := store.AppendBatch(context.Background(), []storage.Record{{EventID: "payload", Payload: payload}}); err != nil {
		t.Fatal(err)
	}
	status, err := store.Status(nil)
	if err != nil {
		t.Fatal(err)
	}
	if status.ByteCount != uint64(len(payload)) {
		t.Fatalf("payload byte count = %d, want %d", status.ByteCount, len(payload))
	}
}

func TestStatusScansRawAndGzipArchivesAfterClose(t *testing.T) {
	root := t.TempDir()
	config := testConfig(root)
	config.SegmentMaxBytes = 1
	config.Compression = true
	store, err := NewWithClock(config, fixedClock{now: time.Unix(100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.AppendBatch(context.Background(), []storage.Record{{EventID: "archived", Timestamp: time.Unix(30, 0)}}); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	status, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.Ready || status.RecordCount != 1 {
		t.Fatalf("closed status = %+v", status)
	}
}

func TestStatusIgnoresUnrelatedFilesAndHonorsCancellation(t *testing.T) {
	root := t.TempDir()
	store, err := NewWithClock(testConfig(root), fixedClock{now: time.Unix(100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := os.WriteFile(filepath.Join(root, "quarantine", "bad.jsonl"), []byte("bad\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "archive", "ignore.tmp"), []byte("bad\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := store.Status(ctx); err == nil {
		t.Fatal("expected cancellation error")
	}
	status, err := store.Status(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if status.RecordCount != 0 {
		t.Fatalf("ignored files counted: %+v", status)
	}
}
