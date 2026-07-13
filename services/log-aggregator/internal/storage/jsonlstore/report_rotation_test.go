package jsonlstore

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func reportTestConfig(root string) Config {
	config := DefaultConfig(root)
	config.SegmentMaxBytes = 1
	config.SegmentMaxAge = time.Hour
	return config
}

func countReportArchives(t *testing.T, root string) int {
	t.Helper()
	count := 0
	_ = filepath.Walk(newReportLayout(root).archiveDir(), func(path string, info os.FileInfo, err error) error {
		if err == nil && info != nil && !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func TestReportStoreRotatesBySizeAndRetrievesArchive(t *testing.T) {
	root := t.TempDir()
	store, err := NewReportStore(reportTestConfig(root))
	if err != nil {
		t.Fatal(err)
	}
	old := testReport(t, "size-old", time.Unix(1, 0).UTC())
	if err := store.Save(context.Background(), old); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "size-new", time.Unix(2, 0).UTC())); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), old.DiagnosticReportID); err != nil {
		t.Fatal(err)
	}
	if countReportArchives(t, root) != 1 {
		t.Fatalf("archives=%d", countReportArchives(t, root))
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportStoreRotatesByAge(t *testing.T) {
	root := t.TempDir()
	now := time.Unix(100, 0).UTC()
	config := reportTestConfig(root)
	config.SegmentMaxBytes = 1 << 20
	config.SegmentMaxAge = time.Minute
	store, err := NewReportStoreWithClock(config, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "age-old", now)); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	if err := store.Save(context.Background(), testReport(t, "age-new", now)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), "age-old"); err != nil {
		t.Fatal(err)
	}
	if countReportArchives(t, root) != 1 {
		t.Fatalf("archives=%d", countReportArchives(t, root))
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportStoreRetrievesCompressedArchive(t *testing.T) {
	root := t.TempDir()
	config := reportTestConfig(root)
	config.Compression = true
	store, err := NewReportStore(config)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "gzip-old", time.Unix(1, 0).UTC())); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "gzip-new", time.Unix(2, 0).UTC())); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), "gzip-old"); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}
