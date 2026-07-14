package logging

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCategoryFilteringHonorsCategoryLevelEnvironment(t *testing.T) {
	resetLogging(t)
	restore := captureStderr(t)
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	t.Setenv(EnvGameLevel, "error")
	Configure("info")

	Game.Info("game filtered")
	Network.Info("network emitted")
	output := restore()

	if strings.Contains(output, "game filtered") {
		t.Fatalf("expected game category info to be filtered, got %q", output)
	}
	if !strings.Contains(output, "network emitted") {
		t.Fatalf("expected network category info to be emitted, got %q", output)
	}
}

func TestConfigureFileOutputReturnsActivePathAndWritesRecords(t *testing.T) {
	resetLogging(t)
	restore := captureStderr(t)
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	Configure("info")

	directory := t.TempDir()
	path, err := ConfigureFileOutput(directory, "game-server")
	if err != nil {
		t.Fatalf("ConfigureFileOutput() error = %v", err)
	}

	wantPath := filepath.Join(directory, "game-server.jsonl.open")
	if path != wantPath {
		t.Fatalf("ConfigureFileOutput() path = %q, want %q", path, wantPath)
	}

	Game.Info("active file output", "mode", "file")
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	stderr := restore()

	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", wantPath, err)
	}
	content := string(data)
	if !strings.Contains(content, `"msg":"active file output"`) {
		t.Fatalf("expected file output to include log message, got %q", content)
	}
	if !strings.Contains(content, `"category":"game"`) {
		t.Fatalf("expected file output to include category field, got %q", content)
	}
	if !strings.Contains(content, `"mode":"file"`) {
		t.Fatalf("expected file output to include payload field, got %q", content)
	}
	if !strings.Contains(stderr, "active file output") {
		t.Fatalf("expected console output while file output was enabled, got %q", stderr)
	}
}

func TestCloseFileOutputReturnsConsoleOnlyLogging(t *testing.T) {
	resetLogging(t)
	restore := captureStderr(t)
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	Configure("info")

	directory := t.TempDir()
	path, err := ConfigureFileOutput(directory, "game-server")
	if err != nil {
		t.Fatalf("ConfigureFileOutput() error = %v", err)
	}

	Game.Info("file phase")
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	Game.Info("console phase")
	stderr := restore()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	content := string(data)
	if !strings.Contains(content, "file phase") {
		t.Fatalf("expected active file to include first record, got %q", content)
	}
	if strings.Contains(content, "console phase") {
		t.Fatalf("expected active file to stop receiving records after CloseFileOutput, got %q", content)
	}
	if !strings.Contains(stderr, "console phase") {
		t.Fatalf("expected console output after CloseFileOutput, got %q", stderr)
	}
}

func resetLogging(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		if err := CloseFileOutput(); err != nil {
			t.Errorf("cleanup CloseFileOutput() error = %v", err)
		}
		Configure("warn")
	})
}

func captureStderr(t *testing.T) func() string {
	t.Helper()

	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = writer

	closed := false
	t.Cleanup(func() {
		if closed {
			return
		}
		os.Stderr = original
		_ = writer.Close()
		_ = reader.Close()
	})

	return func() string {
		t.Helper()
		if !closed {
			closed = true
			os.Stderr = original
			_ = writer.Close()
		}
		data, _ := io.ReadAll(reader)
		_ = reader.Close()
		return string(data)
	}
}
