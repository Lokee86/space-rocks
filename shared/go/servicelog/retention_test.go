package servicelog

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeArchiveFile(t *testing.T, path string, size int, modTime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(path), err)
	}
	data := make([]byte, size)
	for i := range data {
		data[i] = byte('a' + i%26)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("Chtimes(%s) error = %v", path, err)
	}
}

func TestEnforceArchiveRetentionDeletesArchivesOlderThanAge(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	oldPath := archiveSegmentPath(dir, "game-server", now.Add(-4*time.Hour), now.Add(-4*time.Hour), 1)
	newPath := archiveSegmentPath(dir, "game-server", now.Add(-time.Hour), now.Add(-time.Hour), 1)
	ignoredPath := filepath.Join(dir, "archive", "2026", "07", "14", "notes.tmp")
	writeArchiveFile(t, oldPath, 12, now.Add(-4*time.Hour))
	writeArchiveFile(t, newPath, 12, now.Add(-time.Hour))
	writeArchiveFile(t, ignoredPath, 12, now.Add(-6*time.Hour))

	deps := defaultRuntimeDependencies()
	if err := enforceArchiveRetention(FilePolicy{Directory: dir, RetentionMaxAge: 2 * time.Hour, RetentionMaxBytes: 1024}, deps, now); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("old archive still exists: %v", err)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("new archive missing: %v", err)
	}
	if _, err := os.Stat(ignoredPath); err != nil {
		t.Fatalf("ignored file missing: %v", err)
	}
}

func TestEnforceArchiveRetentionDeletesOldestArchivesUntilWithinByteCap(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	base := now.Add(-time.Hour)
	paths := []string{
		archiveSegmentPath(dir, "game-server", base, base, 1),
		archiveSegmentPath(dir, "game-server", base.Add(time.Minute), base.Add(time.Minute), 1),
		archiveSegmentPath(dir, "game-server", base.Add(2*time.Minute), base.Add(2*time.Minute), 1),
	}
	for i, path := range paths {
		stamp := base.Add(time.Duration(i) * time.Minute)
		writeArchiveFile(t, path, 10, stamp)
	}

	deps := defaultRuntimeDependencies()
	if err := enforceArchiveRetention(FilePolicy{Directory: dir, RetentionMaxAge: time.Hour, RetentionMaxBytes: 20}, deps, now); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}

	if _, err := os.Stat(paths[0]); !os.IsNotExist(err) {
		t.Fatalf("oldest archive still exists: %v", err)
	}
	if _, err := os.Stat(paths[1]); err != nil {
		t.Fatalf("middle archive missing: %v", err)
	}
	if _, err := os.Stat(paths[2]); err != nil {
		t.Fatalf("newest archive missing: %v", err)
	}
}

func TestEnforceArchiveRetentionUsesDeterministicOldestFirstOrdering(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	stamp := now.Add(-time.Hour)
	firstPath := archiveSegmentPath(dir, "game-server", stamp, stamp, 1)
	secondPath := archiveSegmentPath(dir, "game-server", stamp, stamp, 2)
	writeArchiveFile(t, firstPath, 8, stamp)
	writeArchiveFile(t, secondPath, 8, stamp)

	deps := defaultRuntimeDependencies()
	if err := enforceArchiveRetention(FilePolicy{Directory: dir, RetentionMaxAge: time.Hour, RetentionMaxBytes: 8}, deps, now); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}

	if _, err := os.Stat(firstPath); !os.IsNotExist(err) {
		t.Fatalf("first archive still exists: %v", err)
	}
	if _, err := os.Stat(secondPath); err != nil {
		t.Fatalf("second archive missing: %v", err)
	}
}

func TestEnforceArchiveRetentionTreatsMissingArchiveDirectoryAsEmpty(t *testing.T) {
	dir := t.TempDir()
	deps := defaultRuntimeDependencies()
	if err := enforceArchiveRetention(FilePolicy{Directory: dir, RetentionMaxAge: time.Hour, RetentionMaxBytes: 8}, deps, time.Now().UTC()); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}
}

func TestEnforceArchiveRetentionIgnoresNonArchiveFiles(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	archivePath := archiveSegmentPath(dir, "game-server", now.Add(-time.Hour), now.Add(-time.Hour), 1)
	ignoredTemp := filepath.Join(dir, "archive", "2026", "07", "14", "orphan.tmp")
	ignoredOpen := filepath.Join(dir, "active", "game-server.jsonl.open")
	writeArchiveFile(t, archivePath, 5, now.Add(-time.Hour))
	writeArchiveFile(t, ignoredTemp, 100, now.Add(-2*time.Hour))
	writeArchiveFile(t, ignoredOpen, 100, now.Add(-2*time.Hour))

	deps := defaultRuntimeDependencies()
	if err := enforceArchiveRetention(FilePolicy{Directory: dir, RetentionMaxAge: time.Hour, RetentionMaxBytes: 10}, deps, now); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}

	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archive file missing: %v", err)
	}
	if _, err := os.Stat(ignoredTemp); err != nil {
		t.Fatalf("ignored temp file missing: %v", err)
	}
	if _, err := os.Stat(ignoredOpen); err != nil {
		t.Fatalf("active file missing: %v", err)
	}
}
