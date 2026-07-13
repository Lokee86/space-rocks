package jsonlstore

import (
	"path/filepath"
	"testing"
	"time"
)

func TestReportLayoutPaths(t *testing.T) {
	layout := newReportLayout("storage")
	if got, want := layout.activePath(), filepath.Join("storage", "reports", "active", "diagnostic-reports.jsonl.open"); got != want {
		t.Fatalf("active path = %q, want %q", got, want)
	}
	if got, want := layout.archiveDir(), filepath.Join("storage", "reports", "archive"); got != want {
		t.Fatalf("archive root = %q, want %q", got, want)
	}
	if got, want := layout.quarantineDir(), filepath.Join("storage", "reports", "quarantine"); got != want {
		t.Fatalf("quarantine root = %q, want %q", got, want)
	}
	archive := layout.archivePath(time.Date(2026, 7, 14, 1, 2, 3, 0, time.UTC), time.Date(2026, 7, 14, 2, 2, 3, 0, time.UTC), 4, true)
	if filepath.Base(archive) != "diagnostic-reports-20260714T010203.000000000Z-20260714T020203.000000000Z-000004.jsonl.gz" {
		t.Fatalf("archive path = %q", archive)
	}
}
