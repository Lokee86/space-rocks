package jsonlstore

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRetentionDeletesExpiredArchivesOnly(t *testing.T) {
	root := t.TempDir()
	layout := NewLayout(root)
	old := filepath.Join(layout.ArchiveDir(), "2026", "01", "01", "old.jsonl")
	keep := filepath.Join(layout.ArchiveDir(), "2026", "01", "01", "keep.jsonl.gz")
	writeRetentionFile(t, old, []byte("old"))
	writeRetentionFile(t, keep, []byte("keep"))
	now := time.Unix(1000, 0)
	_ = os.Chtimes(old, now.Add(-time.Hour), now.Add(-time.Hour))
	_ = os.Chtimes(keep, now, now)
	if err := enforceRetention(layout, now, time.Minute, 1024); err != nil {
		t.Fatal(err)
	}
	assertMissing(t, old)
	assertPresent(t, keep)
}

func TestRetentionDeletesOldestUntilByteCapWithPathTieBreak(t *testing.T) {
	root := t.TempDir()
	layout := NewLayout(root)
	now := time.Unix(1000, 0)
	first := filepath.Join(layout.ArchiveDir(), "a", "first.jsonl")
	second := filepath.Join(layout.ArchiveDir(), "b", "second.jsonl.gz")
	writeRetentionFile(t, first, []byte("12345"))
	writeRetentionFile(t, second, []byte("67890"))
	_ = os.Chtimes(first, now, now)
	_ = os.Chtimes(second, now, now)
	if err := enforceRetention(layout, now, time.Hour, 5); err != nil {
		t.Fatal(err)
	}
	assertMissing(t, first)
	assertPresent(t, second)
}

func TestRetentionPreservesUnrelatedActiveAndQuarantineFiles(t *testing.T) {
	root := t.TempDir()
	layout := NewLayout(root)
	archive := filepath.Join(layout.ArchiveDir(), "2026", "event.txt")
	active := filepath.Join(layout.ActiveDir(), "events.jsonl.open")
	quarantine := filepath.Join(layout.QuarantineDir(), "bad.jsonl")
	for _, path := range []string{archive, active, quarantine} {
		writeRetentionFile(t, path, []byte("keep"))
	}
	if err := enforceRetention(layout, time.Now().Add(24*time.Hour), time.Nanosecond, 1); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{archive, active, quarantine} {
		assertPresent(t, path)
	}
}

func TestRetentionRemovesEmptyArchiveDateDirectories(t *testing.T) {
	root := t.TempDir()
	layout := NewLayout(root)
	archive := filepath.Join(layout.ArchiveDir(), "2026", "07", "14", "expired.jsonl")
	writeRetentionFile(t, archive, []byte("expired"))
	now := time.Unix(1000, 0)
	if err := os.Chtimes(archive, now.Add(-time.Hour), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := enforceRetention(layout, now, time.Minute, 1024); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{
		filepath.Join(layout.ArchiveDir(), "2026", "07", "14"),
		filepath.Join(layout.ArchiveDir(), "2026", "07"),
		filepath.Join(layout.ArchiveDir(), "2026"),
	} {
		assertMissing(t, path)
	}
	assertPresent(t, layout.ArchiveDir())
}

func writeRetentionFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("%s remains: %v", path, err)
	}
}
func assertPresent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s missing: %v", path, err)
	}
}
