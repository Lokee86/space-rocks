package servicelog

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeArchiveCandidate(t *testing.T, directory, relPath string, size int, modTime time.Time) string {
	t.Helper()

	path := filepath.Join(directory, "archive", relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, bytes.Repeat([]byte("x"), size), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("os.Chtimes(%q) error = %v", path, err)
	}
	return path
}

func TestEnforceArchiveRetentionDeletesExpiredArchivesAndIgnoresUnrelatedFiles(t *testing.T) {
	directory := t.TempDir()
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	oldPath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "old.jsonl"), 32, now.Add(-3*time.Hour))
	freshPath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "fresh.jsonl.gz"), 32, now.Add(-30*time.Minute))
	unrelatedPath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "notes.txt"), 16, now.Add(-4*time.Hour))

	if err := enforceArchiveRetention(FilePolicy{Directory: directory, RetentionMaxAge: time.Hour, RetentionMaxBytes: 256}, testRuntimeDependencies(&fakeClock{current: now}), now); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("old archive still exists: %v", err)
	}
	if _, err := os.Stat(freshPath); err != nil {
		t.Fatalf("fresh archive missing: %v", err)
	}
	if _, err := os.Stat(unrelatedPath); err != nil {
		t.Fatalf("unrelated file missing: %v", err)
	}
}

func TestEnforceArchiveRetentionDeletesOldestRemainingUntilByteCap(t *testing.T) {
	directory := t.TempDir()
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	oldestPath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "a.jsonl"), 60, now.Add(-3*time.Hour))
	middlePath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "b.jsonl.gz"), 60, now.Add(-2*time.Hour))
	newestPath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "c.jsonl"), 60, now.Add(-time.Hour))

	if err := enforceArchiveRetention(FilePolicy{Directory: directory, RetentionMaxAge: 24 * time.Hour, RetentionMaxBytes: 120}, testRuntimeDependencies(&fakeClock{current: now}), now); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}

	if _, err := os.Stat(oldestPath); !os.IsNotExist(err) {
		t.Fatalf("oldest archive still exists: %v", err)
	}
	if _, err := os.Stat(middlePath); err != nil {
		t.Fatalf("middle archive missing: %v", err)
	}
	if _, err := os.Stat(newestPath); err != nil {
		t.Fatalf("newest archive missing: %v", err)
	}
}

func TestEnforceArchiveRetentionUsesPathTieBreakerForEqualAges(t *testing.T) {
	directory := t.TempDir()
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	keepPath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "b.jsonl.gz"), 80, now.Add(-time.Hour))
	deletePath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "a.jsonl"), 80, now.Add(-time.Hour))

	if err := enforceArchiveRetention(FilePolicy{Directory: directory, RetentionMaxAge: 24 * time.Hour, RetentionMaxBytes: 80}, testRuntimeDependencies(&fakeClock{current: now}), now); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}

	if _, err := os.Stat(deletePath); !os.IsNotExist(err) {
		t.Fatalf("tie-break archive still exists: %v", err)
	}
	if _, err := os.Stat(keepPath); err != nil {
		t.Fatalf("tie-break survivor missing: %v", err)
	}
}

func TestEnforceArchiveRetentionPreservesActiveFile(t *testing.T) {
	directory := t.TempDir()
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	activePath := filepath.Join(directory, "game-server.jsonl.open")
	if err := os.WriteFile(activePath, []byte("keep me"), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", activePath, err)
	}
	if err := os.Chtimes(activePath, now, now); err != nil {
		t.Fatalf("os.Chtimes(%q) error = %v", activePath, err)
	}
	archivePath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "old.jsonl"), 16, now.Add(-3*time.Hour))

	if err := enforceArchiveRetention(FilePolicy{Directory: directory, RetentionMaxAge: time.Hour, RetentionMaxBytes: 256}, testRuntimeDependencies(&fakeClock{current: now}), now); err != nil {
		t.Fatalf("enforceArchiveRetention() error = %v", err)
	}

	data, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", activePath, err)
	}
	if string(data) != "keep me" {
		t.Fatalf("active file contents = %q, want keep me", string(data))
	}
	if _, err := os.Stat(archivePath); !os.IsNotExist(err) {
		t.Fatalf("old archive still exists: %v", err)
	}
}

func TestOpenRunsArchiveRetentionOnStartup(t *testing.T) {
	directory := t.TempDir()
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	oldPath := writeArchiveCandidate(t, directory, filepath.Join("2026-07-14", "old.jsonl"), 32, now.Add(-3*time.Hour))
	clock := &fakeClock{current: now}

	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: func() FilePolicy {
			cfg := validFileConfig(directory)
			cfg.RetentionMaxAge = time.Hour
			cfg.RetentionMaxBytes = 256
			return cfg
		}(),
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, testRuntimeDependencies(clock))
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer runtime.Close()

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("stale archive still exists after Open: %v", err)
	}
	activePath := filepath.Join(directory, "game-server.jsonl.open")
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("active file missing after Open: %v", err)
	}
}
