package jsonlstore

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage"
)

type advancingClock struct{ now time.Time }

func (clock *advancingClock) Now() time.Time { return clock.now }

func rotationConfig(root string) Config {
	config := testConfig(root)
	config.SegmentMaxBytes = 80
	config.SegmentMaxAge = time.Hour
	config.Compression = false
	return config
}

func TestAppendRotatesBySizeAndCloseFinalizes(t *testing.T) {
	root := t.TempDir()
	config := rotationConfig(root)
	store, err := NewWithClock(config, fixedClock{now: time.Unix(100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if err = store.AppendBatch(context.Background(), []storage.Record{{EventID: strings.Repeat("x", 50)}, {EventID: "second"}}); err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	var matches []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && strings.HasSuffix(path, ".jsonl") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("expected finalized archive")
	}
}

func TestAppendRotatesByAge(t *testing.T) {
	root := t.TempDir()
	clock := &advancingClock{now: time.Unix(100, 0)}
	config := rotationConfig(root)
	config.SegmentMaxBytes = 1024
	config.SegmentMaxAge = time.Second
	store, err := NewWithClock(config, clock)
	if err != nil {
		t.Fatal(err)
	}
	if err = store.AppendBatch(context.Background(), []storage.Record{{EventID: "first"}}); err != nil {
		t.Fatal(err)
	}
	clock.now = clock.now.Add(2 * time.Second)
	if err = store.AppendBatch(context.Background(), []storage.Record{{EventID: "second"}}); err != nil {
		t.Fatal(err)
	}
	if store.activeBytes == 0 {
		t.Fatal("active segment was not reopened")
	}
	_ = store.Close()
}

func TestOversizedRecordIsFinalizedImmediately(t *testing.T) {
	root := t.TempDir()
	config := rotationConfig(root)
	config.SegmentMaxBytes = 20
	store, err := NewWithClock(config, fixedClock{now: time.Unix(100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if err = store.AppendBatch(context.Background(), []storage.Record{{EventID: strings.Repeat("x", 100)}}); err != nil {
		t.Fatal(err)
	}
	if store.activeBytes == 0 {
		t.Fatal("oversized batch was not kept intact in the active segment")
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestCompressedRotationUsesGzipArchive(t *testing.T) {
	root := t.TempDir()
	config := rotationConfig(root)
	config.Compression = true
	store, err := NewWithClock(config, fixedClock{now: time.Unix(100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if err = store.AppendBatch(context.Background(), []storage.Record{{EventID: strings.Repeat("x", 100)}}); err != nil {
		t.Fatal(err)
	}
	if err = store.Close(); err != nil {
		t.Fatal(err)
	}
	var matches []string
	err = filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() && strings.HasSuffix(path, ".gz") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("gzip archives = %v", matches)
	}
	if _, err = os.Stat(matches[0]); err != nil {
		t.Fatal(err)
	}
}
