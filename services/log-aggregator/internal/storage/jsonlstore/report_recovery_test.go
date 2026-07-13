package jsonlstore

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

func writeInterruptedReports(t *testing.T, root string, lines ...string) string {
	t.Helper()
	path := newReportLayout(root).activePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func reportJSON(t *testing.T, report storage.Report) string {
	t.Helper()
	return string(report.RawJSON)
}

func TestReportRecoveryTruncatesTrailingLineAndArchivesValidReports(t *testing.T) {
	root := t.TempDir()
	valid := testReport(t, "recovered-valid", time.Unix(1, 0).UTC())
	writeInterruptedReports(t, root, reportJSON(t, valid), "{partial")
	store, err := NewReportStore(DefaultConfig(root))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), valid.DiagnosticReportID); err != nil {
		t.Fatal(err)
	}
	if countReportArchives(t, root) != 1 {
		t.Fatalf("archives=%d", countReportArchives(t, root))
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportRecoveryFinalizesValidActiveFile(t *testing.T) {
	root := t.TempDir()
	valid := testReport(t, "valid-interrupted", time.Unix(2, 0).UTC())
	writeInterruptedReports(t, root, reportJSON(t, valid)+"\n")
	store, err := NewReportStore(DefaultConfig(root))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), valid.DiagnosticReportID); err != nil {
		t.Fatal(err)
	}
	if countReportArchives(t, root) != 1 {
		t.Fatalf("archives=%d", countReportArchives(t, root))
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportRecoveryQuarantinesMalformedActiveFile(t *testing.T) {
	root := t.TempDir()
	valid := testReport(t, "valid-before-malformed", time.Unix(3, 0).UTC())
	writeInterruptedReports(t, root, reportJSON(t, valid), "{malformed", reportJSON(t, testReport(t, "after", time.Unix(4, 0).UTC())))
	store, err := NewReportStore(DefaultConfig(root))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(newReportLayout(root).activePath()); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(context.Background(), testReport(t, "fresh", time.Unix(5, 0).UTC())); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(newReportLayout(root).quarantineDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("quarantine entries=%d", len(entries))
	}
}

func TestReportRecoveryQuarantinesValidJSONMissingProjectionFields(t *testing.T) {
	root := t.TempDir()
	writeInterruptedReports(t, root, `{"created_at":"1970-01-01T00:00:03Z"}`)
	store, err := NewReportStore(DefaultConfig(root))
	if err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(newReportLayout(root).quarantineDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("quarantine entries=%d", len(entries))
	}
	if err := store.Save(context.Background(), testReport(t, "fresh-after-projection-error", time.Unix(5, 0).UTC())); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestReportRecoveryFinalizesUnterminatedValidReport(t *testing.T) {
	root := t.TempDir()
	valid := testReport(t, "unterminated-valid", time.Unix(6, 0).UTC())
	writeInterruptedReports(t, root, reportJSON(t, valid))
	store, err := NewReportStore(DefaultConfig(root))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get(context.Background(), valid.DiagnosticReportID); err != nil {
		t.Fatal(err)
	}
	if countReportArchives(t, root) != 1 {
		t.Fatalf("archives=%d", countReportArchives(t, root))
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
}
