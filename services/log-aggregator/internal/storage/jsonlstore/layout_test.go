package jsonlstore

import (
	"path/filepath"
	"testing"
	"time"
)

func TestLayoutPaths(t *testing.T) {
	layout := NewLayout("storage")
	if got, want := layout.ActivePath(), filepath.Join("storage", "active", "events.jsonl.open"); got != want {
		t.Fatalf("active path = %q, want %q", got, want)
	}
	if got, want := layout.ArchiveDir(), filepath.Join("storage", "archive"); got != want {
		t.Fatalf("archive dir = %q, want %q", got, want)
	}
	if got, want := layout.QuarantineDir(), filepath.Join("storage", "quarantine"); got != want {
		t.Fatalf("quarantine dir = %q, want %q", got, want)
	}
}

func TestArchivePathUsesUTCStartDateAndDeterministicName(t *testing.T) {
	location := time.FixedZone("offset", -5*60*60)
	start := time.Date(2026, 7, 13, 23, 30, 1, 123456789, location)
	end := start.Add(2 * time.Minute)
	layout := NewLayout("storage")
	want := filepath.Join(
		"storage", "archive", "2026", "07", "14",
		"events-20260714T043001.123456789Z-20260714T043201.123456789Z-000007.jsonl",
	)
	if got := layout.ArchivePath(start, end, 7, false); got != want {
		t.Fatalf("archive path = %q, want %q", got, want)
	}
	if got := layout.ArchivePath(start, end, 7, false); got != want {
		t.Fatalf("archive path changed between calls: %q", got)
	}
}

func TestArchivePathCompressionSuffix(t *testing.T) {
	start := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	end := start.Add(time.Hour)
	path := NewLayout("storage").ArchivePath(start, end, 1, true)
	want := filepath.Join("storage", "archive", "2026", "01", "02", "events-20260102T030405.000000000Z-20260102T040405.000000000Z-000001.jsonl.gz")
	if path != want {
		t.Fatalf("compressed archive path = %q, want %q", path, want)
	}
}
