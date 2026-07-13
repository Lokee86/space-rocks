package jsonlstore

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFinalizeArchiveRenamesRawSource(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "active.jsonl.open")
	destination := filepath.Join(root, "archive", "events.jsonl")
	writeArchiveSource(t, source, "one\ntwo\n")

	if err := finalizeArchive(source, destination, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists, error = %v", err)
	}
	if got := readArchiveFile(t, destination); got != "one\ntwo\n" {
		t.Fatalf("archive content = %q", got)
	}
}

func TestFinalizeArchiveWritesValidGzipAndRemovesSource(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "active.jsonl.open")
	destination := filepath.Join(root, "archive", "events.jsonl.gz")
	writeArchiveSource(t, source, "compressed\n")

	if err := finalizeArchive(source, destination, true); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(destination)
	if err != nil {
		t.Fatal(err)
	}
	reader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	_ = reader.Close()
	_ = file.Close()
	if string(data) != "compressed\n" {
		t.Fatalf("decompressed archive = %q", data)
	}
	if _, err := os.Stat(source); !os.IsNotExist(err) {
		t.Fatalf("source still exists, error = %v", err)
	}
}

func TestFinalizeArchiveCleansTemporaryOutputOnFailure(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "missing.jsonl.open")
	destination := filepath.Join(root, "archive", "events.jsonl.gz")
	if err := finalizeArchive(source, destination, true); err == nil {
		t.Fatal("expected missing source error")
	}
	entries, err := os.ReadDir(filepath.Dir(destination))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary files remain: %v", entries)
	}
}

func TestFinalizeArchivePreservesSourceWhenDestinationCannotBeCreated(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "active.jsonl.open")
	blocked := filepath.Join(root, "blocked")
	writeArchiveSource(t, source, "preserve\n")
	if err := os.WriteFile(blocked, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := finalizeArchive(source, filepath.Join(blocked, "events.jsonl"), false); err == nil {
		t.Fatal("expected destination error")
	}
	if got := readArchiveFile(t, source); got != "preserve\n" {
		t.Fatalf("source content = %q", got)
	}
}

func writeArchiveSource(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readArchiveFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(path, ".gz") {
		t.Fatal("readArchiveFile is for raw files")
	}
	return string(data)
}
