package filestore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/audit"
)

const recordID = "123e4567-e89b-12d3-a456-426614174010"

func testRecord() audit.Record { return audit.Record{Version: audit.RecordVersion, AuditEventID: recordID, SourceEventID: "123e4567-e89b-12d3-a456-426614174011", Payload: json.RawMessage(`{"ok":true}`)} }

func TestStoreRoundTripAndDuplicate(t *testing.T) {
	store, err := New(t.TempDir(), 4096)
	if err != nil { t.Fatal(err) }
	record := testRecord()
	if err := store.Save(context.Background(), record); err != nil { t.Fatal(err) }
	got, err := store.Get(context.Background(), recordID)
	if err != nil || string(got.Payload) != string(record.Payload) { t.Fatalf("got %#v, err %v", got, err) }
	if err := store.Save(context.Background(), record); !errors.Is(err, ErrDuplicateRecord) { t.Fatalf("got %v", err) }
}

func TestStoreConcurrentSavesAndCleanup(t *testing.T) {
	store, err := New(t.TempDir(), 4096)
	if err != nil { t.Fatal(err) }
	var wait sync.WaitGroup
	results := make(chan error, 2)
	for range 2 { wait.Add(1); go func() { defer wait.Done(); results <- store.Save(context.Background(), testRecord()) }() }
	wait.Wait(); close(results)
	var success, duplicate int
	for err := range results { if err == nil { success++ } else if errors.Is(err, ErrDuplicateRecord) { duplicate++ } else { t.Fatal(err) } }
	if success != 1 || duplicate != 1 { t.Fatalf("success=%d duplicate=%d", success, duplicate) }
	if _, err := store.Get(context.Background(), recordID); err != nil { t.Fatal(err) }
	if files, _ := filepath.Glob(filepath.Join(store.root, ".audit-*.tmp")); len(files) != 0 { t.Fatalf("temporary files remain: %v", files) }
}

func TestStoreValidationAndLimits(t *testing.T) {
	store, err := New(t.TempDir(), 32)
	if err != nil { t.Fatal(err) }
	if err := store.Save(context.Background(), audit.Record{AuditEventID: "bad"}); !errors.Is(err, audit.ErrInvalidAuditEventID) { t.Fatal(err) }
	if _, err := store.Get(context.Background(), "bad"); !errors.Is(err, audit.ErrInvalidAuditEventID) { t.Fatal(err) }
	if _, err := store.Get(context.Background(), recordID); !errors.Is(err, audit.ErrAuditRecordNotFound) { t.Fatal(err) }
	if err := store.Save(context.Background(), testRecord()); !errors.Is(err, ErrRecordTooLarge) { t.Fatal(err) }
	path := filepath.Join(store.root, recordID+".json")
	if err := os.WriteFile(path, []byte("{bad"), 0o644); err != nil { t.Fatal(err) }
	if _, err := store.Get(context.Background(), recordID); !errors.Is(err, ErrInvalidStoredData) { t.Fatal(err) }
	if err := os.WriteFile(path, []byte(`{"audit_event_id":"123e4567-e89b-12d3-a456-426614174012"}`), 0o644); err != nil { t.Fatal(err) }
	if _, err := store.Get(context.Background(), recordID); !errors.Is(err, ErrInvalidStoredData) { t.Fatal(err) }
	if err := os.WriteFile(path, make([]byte, 33), 0o644); err != nil { t.Fatal(err) }
	if _, err := store.Get(context.Background(), recordID); !errors.Is(err, ErrRecordTooLarge) { t.Fatal(err) }
}

func TestStoreCanceledAndRequiredRoot(t *testing.T) {
	if _, err := New("", 0); !errors.Is(err, ErrInvalidRoot) { t.Fatal(err) }
	store, err := New(t.TempDir(), 4096)
	if err != nil { t.Fatal(err) }
	ctx, cancel := context.WithCancel(context.Background()); cancel()
	if err := store.Save(ctx, testRecord()); !errors.Is(err, context.Canceled) { t.Fatal(err) }
	if _, err := store.Get(ctx, recordID); !errors.Is(err, context.Canceled) { t.Fatal(err) }
}
