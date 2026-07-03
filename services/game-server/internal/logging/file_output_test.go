package logging

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenSequentialLogFileCreatesFirstFile(t *testing.T) {
	baseDir := filepath.Join(t.TempDir(), "logs")

	file, path, err := openSequentialLogFile(baseDir, "server")
	if err != nil {
		t.Fatalf("openSequentialLogFile returned error: %v", err)
	}
	defer file.Close()

	expectedPath := filepath.Join(baseDir, "server-000001.jsonl")
	if path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, path)
	}

	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected file to exist at %q: %v", expectedPath, err)
	}
}

func TestOpenSequentialLogFileCreatesSecondFileWhenFirstExists(t *testing.T) {
	baseDir := t.TempDir()
	firstPath := filepath.Join(baseDir, "server-000001.jsonl")
	if err := os.WriteFile(firstPath, []byte("existing"), 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	file, path, err := openSequentialLogFile(baseDir, "server")
	if err != nil {
		t.Fatalf("openSequentialLogFile returned error: %v", err)
	}
	defer file.Close()

	expectedPath := filepath.Join(baseDir, "server-000002.jsonl")
	if path != expectedPath {
		t.Fatalf("expected path %q, got %q", expectedPath, path)
	}

	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected file to exist at %q: %v", expectedPath, err)
	}
}

func TestOpenSequentialLogFileUsesSixDigitNumbering(t *testing.T) {
	baseDir := t.TempDir()

	file, path, err := openSequentialLogFile(baseDir, "server")
	if err != nil {
		t.Fatalf("openSequentialLogFile returned error: %v", err)
	}
	defer file.Close()

	expectedName := "server-000001.jsonl"
	if filepath.Base(path) != expectedName {
		t.Fatalf("expected filename %q, got %q", expectedName, filepath.Base(path))
	}
}
