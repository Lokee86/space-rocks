package jsonlstore

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

func TestRecoveryMissingAndEmptyActive(t *testing.T) {
	root := t.TempDir()
	layout := testRecoveryLayout(t, root)
	if _, _, err := recoverActive(layout, fixedClock{now: time.Unix(100, 0)}, testConfig(root), 0); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(layout.ActivePath(), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := recoverActive(layout, fixedClock{now: time.Unix(100, 0)}, testConfig(root), 0); err != nil {
		t.Fatal(err)
	}
}

func TestRecoveryArchivesCleanActive(t *testing.T) {
	root := t.TempDir()
	layout := testRecoveryLayout(t, root)
	line, _ := EncodeRecord(storage.Record{EventID: "clean"})
	if err := os.WriteFile(layout.ActivePath(), line, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := recoverActive(layout, fixedClock{now: time.Unix(200, 0)}, testConfig(root), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(layout.ActivePath()); !os.IsNotExist(err) {
		t.Fatalf("active file remains: %v", err)
	}
}

func TestRecoveryDropsOnlyTruncatedFinalLine(t *testing.T) {
	root := t.TempDir()
	layout := testRecoveryLayout(t, root)
	config := testConfig(root)
	config.Compression = false
	start := time.Unix(100, 0)
	valid, _ := EncodeRecord(storage.Record{EventID: "valid"})
	data := append(valid, []byte(`{"event_id":"truncated`)...)
	if err := os.WriteFile(layout.ActivePath(), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(layout.ActivePath(), start, start); err != nil {
		t.Fatal(err)
	}
	if _, _, err := recoverActive(layout, fixedClock{now: time.Unix(200, 0)}, config, 0); err != nil {
		t.Fatal(err)
	}
	archives := findRecoveryFiles(t, layout.ArchiveDir(), ".jsonl")
	if len(archives) != 1 {
		t.Fatalf("raw archives = %v", archives)
	}
	file, err := os.Open(archives[0])
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(file)
	var records []storage.Record
	for scanner.Scan() {
		record, decodeErr := DecodeRecord(scanner.Bytes())
		if decodeErr != nil {
			t.Fatal(decodeErr)
		}
		records = append(records, record)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].EventID != "valid" {
		t.Fatalf("recovered records = %+v", records)
	}
}

func TestRecoveryQuarantinesEarlierCorruptionAndContinuesDegraded(t *testing.T) {
	root := t.TempDir()
	layout := testRecoveryLayout(t, root)
	valid, _ := EncodeRecord(storage.Record{EventID: "valid"})
	data := append([]byte(`{"event_id":`), '\n')
	data = append(data, valid...)
	if err := os.WriteFile(layout.ActivePath(), data, 0o644); err != nil {
		t.Fatal(err)
	}
	store, err := NewWithClock(testConfig(root), fixedClock{now: time.Unix(200, 0)})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if !store.degraded {
		t.Fatal("expected degraded recovery state")
	}
	if err := store.AppendBatch(context.Background(), []storage.Record{{EventID: "healthy"}}); err != nil {
		t.Fatal(err)
	}
	if err := store.Flush(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(layout.QuarantineDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || !strings.Contains(entries[0].Name(), "events.jsonl.open") {
		t.Fatalf("quarantine entries = %v", entries)
	}
}

func TestRecoveryUsesNextArchiveSequence(t *testing.T) {
	root := t.TempDir()
	layout := testRecoveryLayout(t, root)
	config := testConfig(root)
	config.Compression = false
	start := time.Unix(100, 0)
	end := time.Unix(200, 0)
	line, _ := EncodeRecord(storage.Record{EventID: "restart"})
	if err := os.WriteFile(layout.ActivePath(), line, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(layout.ActivePath(), start, start); err != nil {
		t.Fatal(err)
	}
	sequenceZero := layout.ArchivePath(start, end, 0, false)
	if err := os.MkdirAll(filepath.Dir(sequenceZero), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sequenceZero, []byte("existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	next, _, err := recoverActive(layout, fixedClock{now: end}, config, 0)
	if err != nil {
		t.Fatal(err)
	}
	if next != 2 {
		t.Fatalf("next archive sequence = %d, want 2", next)
	}
	if _, err := os.Stat(layout.ArchivePath(start, end, 1, false)); err != nil {
		t.Fatalf("sequence-1 archive missing: %v", err)
	}
}

func findRecoveryFiles(t *testing.T, root, suffix string) []string {
	t.Helper()
	var paths []string
	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, suffix) {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return paths
}

func testRecoveryLayout(t *testing.T, root string) Layout {
	t.Helper()
	layout := NewLayout(root)
	if err := os.MkdirAll(layout.ActiveDir(), 0o755); err != nil {
		t.Fatal(err)
	}
	return layout
}
