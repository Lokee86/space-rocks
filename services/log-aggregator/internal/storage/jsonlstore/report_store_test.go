package jsonlstore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

func testReport(t *testing.T, id string, created time.Time) storage.Report {
	t.Helper()
	data, err := json.Marshal(struct {
		ID        string    `json:"diagnostic_report_id"`
		CreatedAt time.Time `json:"created_at"`
	}{id, created})
	if err != nil {
		t.Fatal(err)
	}
	return storage.Report{DiagnosticReportID: id, CreatedAt: created, RawJSON: data}
}

func TestReportStoreSaveGetRestartAndNotFound(t *testing.T) {
	root := t.TempDir()
	config := DefaultConfig(root)
	store, err := NewReportStore(config)
	if err != nil {
		t.Fatal(err)
	}
	want := testReport(t, "550e8400-e29b-41d4-a716-446655440000", time.Unix(10, 0).UTC())
	if err := store.Save(context.Background(), want); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	store, err = NewReportStore(config)
	if err != nil {
		t.Fatal(err)
	}
	got, err := store.Get(context.Background(), want.DiagnosticReportID)
	if err != nil {
		t.Fatal(err)
	}
	if got.DiagnosticReportID != want.DiagnosticReportID || !got.CreatedAt.Equal(want.CreatedAt) {
		t.Fatalf("got %#v", got)
	}
	if _, err := store.Get(context.Background(), "missing"); !errors.Is(err, ErrReportNotFound) {
		t.Fatalf("err=%v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportStoreDeleteExpiredReturnsCount(t *testing.T) {
	store, err := NewReportStore(DefaultConfig(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "old", time.Unix(10, 0).UTC())); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "keep", time.Unix(20, 0).UTC())); err != nil {
		t.Fatal(err)
	}
	removed, err := store.DeleteExpired(context.Background(), time.Unix(20, 0).UTC())
	if err != nil || removed != 1 {
		t.Fatalf("removed=%d err=%v", removed, err)
	}
	if _, err := store.Get(context.Background(), "old"); !errors.Is(err, ErrReportNotFound) {
		t.Fatalf("old err=%v", err)
	}
	if _, err := store.Get(context.Background(), "keep"); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportStoreIgnoresMalformedFinalLine(t *testing.T) {
	root := t.TempDir()
	config := DefaultConfig(root)
	want := testReport(t, "valid", time.Unix(10, 0).UTC())
	line, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	path := newReportLayout(root).activePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(append(line, '\n'), []byte("{truncated")...), 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := NewReportStore(config)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), want.DiagnosticReportID); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportStoreEnforceRetentionUsesConfiguredClock(t *testing.T) {
	config := DefaultConfig(t.TempDir())
	now := time.Unix(20*24*60*60, 0).UTC()
	store, err := NewReportStoreWithClock(config, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "expired", now.Add(-15*24*time.Hour))); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "current", now.Add(-13*24*time.Hour))); err != nil {
		t.Fatal(err)
	}
	removed, err := store.EnforceRetention(context.Background())
	if err != nil || removed != 1 {
		t.Fatalf("removed=%d err=%v", removed, err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportStoreClose(t *testing.T) {
	store, err := NewReportStore(DefaultConfig(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), storage.Report{}); !errors.Is(err, storage.ErrClosed) {
		t.Fatalf("err=%v", err)
	}
}
