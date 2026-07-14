package servicelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeInterruptedActiveSegment(t *testing.T, directory string, contents string, modTime time.Time) string {
	t.Helper()

	path := filepath.Join(directory, "game-server.jsonl.open")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("os.MkdirAll(%q) error = %v", directory, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("os.WriteFile(%q) error = %v", path, err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("os.Chtimes(%q) error = %v", path, err)
	}
	return path
}

func openRecoveredRuntime(t *testing.T, directory string, compression bool, now time.Time) *Runtime {
	t.Helper()

	clock := &fakeClock{current: now}
	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           func() FilePolicy { cfg := validFileConfig(directory); cfg.CompressionEnabled = compression; return cfg }(),
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, testRuntimeDependencies(clock))
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	return runtime
}

func TestOpenRecoversCompleteActiveSegmentIntoArchive(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	writeInterruptedActiveSegment(t, directory, "{\"msg\":\"one\"}\n{\"msg\":\"two\"}\n", now.Add(-2*time.Hour))

	runtime := openRecoveredRuntime(t, directory, false, now)
	defer runtime.Close()

	paths := collectLogFiles(t, directory)
	if len(paths) != 2 {
		t.Fatalf("log file count = %d, want 2; paths = %v", len(paths), paths)
	}
	activePath := filepath.Join(directory, "game-server.jsonl.open")
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("os.Stat(%q) error = %v", activePath, err)
	}
	if info, err := os.Stat(activePath); err != nil || info.Size() != 0 {
		t.Fatalf("active file size = %d, err = %v, want 0 and nil", info.Size(), err)
	}

	var archivePath string
	for _, path := range paths {
		if path != activePath {
			archivePath = path
		}
	}
	if archivePath == "" {
		t.Fatal("archive path not found")
	}
	if strings.HasSuffix(archivePath, ".gz") {
		t.Fatalf("archive path = %q, want uncompressed archive", archivePath)
	}
	assertArchiveFilename(t, archivePath, "game-server")

	records := readJSONLines(t, archivePath)
	if len(records) != 2 {
		t.Fatalf("recovered record count = %d, want 2", len(records))
	}
	if got := records[0]["msg"]; got != "one" {
		t.Fatalf("first recovered message = %v, want one", got)
	}
	if got := records[1]["msg"]; got != "two" {
		t.Fatalf("second recovered message = %v, want two", got)
	}
}

func TestOpenDropsTruncatedFinalLineDuringRecovery(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	writeInterruptedActiveSegment(t, directory, "{\"msg\":\"one\"}\n{\"msg\":\"two\"}", now.Add(-2*time.Hour))

	runtime := openRecoveredRuntime(t, directory, false, now)
	defer runtime.Close()

	paths := collectLogFiles(t, directory)
	if len(paths) != 2 {
		t.Fatalf("log file count = %d, want 2; paths = %v", len(paths), paths)
	}
	activePath := filepath.Join(directory, "game-server.jsonl.open")
	var archivePath string
	for _, path := range paths {
		if path != activePath {
			archivePath = path
		}
	}
	if archivePath == "" {
		t.Fatal("archive path not found")
	}
	records := readJSONLines(t, archivePath)
	if len(records) != 1 {
		t.Fatalf("recovered record count = %d, want 1", len(records))
	}
	if got := records[0]["msg"]; got != "one" {
		t.Fatalf("recovered message = %v, want one", got)
	}
	if _, err := os.Stat(activePath); err != nil {
		t.Fatalf("active file missing: %v", err)
	}
	if info, err := os.Stat(activePath); err != nil || info.Size() != 0 {
		t.Fatalf("active file size = %d, err = %v, want 0 and nil", info.Size(), err)
	}
}

func TestOpenRemovesEmptyOrFullyTruncatedActiveSegment(t *testing.T) {
	cases := []struct {
		name     string
		contents string
	}{
		{name: "empty", contents: ""},
		{name: "truncated-only", contents: "{\"msg\":\"one\"}"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			directory := filepath.Join(t.TempDir(), "logs")
			now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
			writeInterruptedActiveSegment(t, directory, tc.contents, now.Add(-2*time.Hour))

			runtime := openRecoveredRuntime(t, directory, false, now)
			defer runtime.Close()

			paths := collectLogFiles(t, directory)
			if len(paths) != 1 {
				t.Fatalf("log file count = %d, want 1; paths = %v", len(paths), paths)
			}
			activePath := filepath.Join(directory, "game-server.jsonl.open")
			if paths[0] != activePath {
				t.Fatalf("unexpected file path = %q, want %q", paths[0], activePath)
			}
			if _, err := os.Stat(filepath.Join(directory, "archive")); !os.IsNotExist(err) {
				t.Fatalf("archive directory presence = %v, want not exist", err)
			}
			if info, err := os.Stat(activePath); err != nil || info.Size() != 0 {
				t.Fatalf("active file size = %d, err = %v, want 0 and nil", info.Size(), err)
			}
		})
	}
}

func TestOpenRecoversCompressedActiveSegment(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	writeInterruptedActiveSegment(t, directory, "{\"msg\":\"one\"}\n{\"msg\":\"two\"}\n", now.Add(-2*time.Hour))

	runtime := openRecoveredRuntime(t, directory, true, now)
	defer runtime.Close()

	paths := collectLogFiles(t, directory)
	if len(paths) != 2 {
		t.Fatalf("log file count = %d, want 2; paths = %v", len(paths), paths)
	}
	activePath := filepath.Join(directory, "game-server.jsonl.open")
	var archivePath string
	for _, path := range paths {
		if path != activePath {
			archivePath = path
		}
	}
	if archivePath == "" {
		t.Fatal("archive path not found")
	}
	if !strings.HasSuffix(archivePath, ".jsonl.gz") {
		t.Fatalf("archive path = %q, want compressed archive", archivePath)
	}
	assertArchiveFilename(t, strings.TrimSuffix(archivePath, ".gz"), "game-server")
	records := readGzipJSONLines(t, archivePath)
	if len(records) != 2 {
		t.Fatalf("recovered record count = %d, want 2", len(records))
	}
	if got := records[0]["msg"]; got != "one" {
		t.Fatalf("first recovered message = %v, want one", got)
	}
	if got := records[1]["msg"]; got != "two" {
		t.Fatalf("second recovered message = %v, want two", got)
	}
	if info, err := os.Stat(activePath); err != nil || info.Size() != 0 {
		t.Fatalf("active file size = %d, err = %v, want 0 and nil", info.Size(), err)
	}
	if _, err := os.Stat(strings.TrimSuffix(archivePath, ".gz")); !os.IsNotExist(err) {
		t.Fatalf("uncompressed recovered archive source still exists: %v", err)
	}
}
