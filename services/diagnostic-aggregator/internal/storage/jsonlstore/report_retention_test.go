package jsonlstore

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

func saveRetentionReports(t *testing.T, store *ReportStore, reports ...storage.Report) {
	t.Helper()
	for _, report := range reports {
		if err := store.Save(context.Background(), report); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReportRetentionRotatesActiveAndRemovesExpiredAcrossArchives(t *testing.T) {
	root := t.TempDir()
	config := reportTestConfig(root)
	store, err := NewReportStore(config)
	if err != nil {
		t.Fatal(err)
	}
	saveRetentionReports(t, store,
		testReport(t, "expired-one", time.Unix(1, 0).UTC()),
		testReport(t, "current-one", time.Unix(20, 0).UTC()),
		testReport(t, "expired-two", time.Unix(2, 0).UTC()),
		testReport(t, "current-two", time.Unix(21, 0).UTC()))
	removed, err := store.DeleteExpired(context.Background(), time.Unix(10, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	if removed != 2 {
		t.Fatalf("removed=%d, want 2", removed)
	}
	for _, id := range []string{"current-one", "current-two"} {
		if _, err := store.Get(context.Background(), id); err != nil {
			t.Fatalf("get %s: %v", id, err)
		}
	}
	for _, id := range []string{"expired-one", "expired-two"} {
		if _, err := store.Get(context.Background(), id); !errors.Is(err, ErrReportNotFound) {
			t.Fatalf("get %s: %v", id, err)
		}
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportRetentionCompressedArchivesRemainReadable(t *testing.T) {
	root := t.TempDir()
	config := reportTestConfig(root)
	config.Compression = true
	store, err := NewReportStore(config)
	if err != nil {
		t.Fatal(err)
	}
	saveRetentionReports(t, store, testReport(t, "gzip-expired", time.Unix(1, 0).UTC()), testReport(t, "gzip-current", time.Unix(20, 0).UTC()))
	removed, err := store.DeleteExpired(context.Background(), time.Unix(10, 0).UTC())
	if err != nil || removed != 1 {
		t.Fatalf("removed=%d err=%v", removed, err)
	}
	if _, err := store.Get(context.Background(), "gzip-current"); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportRetentionDeletesEmptyArchivesAndDatedDirectories(t *testing.T) {
	root := t.TempDir()
	config := reportTestConfig(root)
	store, err := NewReportStore(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "only-expired", time.Unix(1, 0).UTC())); err != nil {
		t.Fatal(err)
	}
	removed, err := store.DeleteExpired(context.Background(), time.Unix(10, 0).UTC())
	if err != nil || removed != 1 {
		t.Fatalf("removed=%d err=%v", removed, err)
	}
	archiveRoot := newReportLayout(root).archiveDir()
	entries, err := os.ReadDir(archiveRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("archive entries=%d, want empty", len(entries))
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportStoreEnforceRetentionUsesFourteenDayDefault(t *testing.T) {
	root := t.TempDir()
	config := DefaultConfig(root)
	now := time.Unix(30*24*60*60, 0).UTC()
	store, err := NewReportStoreWithClock(config, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "expired-default", now.Add(-15*24*time.Hour))); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "current-default", now.Add(-13*24*time.Hour))); err != nil {
		t.Fatal(err)
	}
	removed, err := store.EnforceRetention(context.Background())
	if err != nil || removed != 1 {
		t.Fatalf("removed=%d err=%v", removed, err)
	}
	if _, err := store.Get(context.Background(), "current-default"); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}
