package jsonlstore

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/storage"
)

type fixedClock struct{ now time.Time }

func (clock fixedClock) Now() time.Time { return clock.now }

func testConfig(root string) Config {
	config := DefaultConfig(root)
	config.FlushInterval = 0
	return config
}

func TestAppendBatchPreservesOrderAndFlushes(t *testing.T) {
	root := t.TempDir()
	store, err := NewWithClock(testConfig(root), fixedClock{now: time.Unix(100, 0)})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	records := []storage.Record{{EventID: "one"}, {EventID: "two"}, {EventID: "three"}}
	if err := store.AppendBatch(context.Background(), records); err != nil {
		t.Fatal(err)
	}
	if err := store.Flush(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "active", "events.jsonl.open"))
	if err != nil {
		t.Fatal(err)
	}
	for index, want := range []string{"one", "two", "three"} {
		if got, err := DecodeRecord(bytesLine(data, index)); err != nil || got.EventID != want {
			t.Fatalf("line %d = %q, error = %v", index, got.EventID, err)
		}
	}
	if store.activeBytes != int64(len(data)) || store.segmentStart != time.Unix(100, 0) {
		t.Fatalf("active tracking = bytes %d, start %v", store.activeBytes, store.segmentStart)
	}
}

func TestAppendBatchRespectsCanceledContext(t *testing.T) {
	store, err := NewWithClock(testConfig(t.TempDir()), fixedClock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := store.AppendBatch(ctx, []storage.Record{{EventID: "ignored"}}); err != context.Canceled {
		t.Fatalf("AppendBatch() error = %v, want context canceled", err)
	}
	if store.activeBytes != 0 {
		t.Fatalf("canceled append changed active bytes: %d", store.activeBytes)
	}
}

func TestAppendBatchConcurrentWriters(t *testing.T) {
	store, err := NewWithClock(testConfig(t.TempDir()), fixedClock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	const writers = 8
	const recordsPerWriter = 25
	var group sync.WaitGroup
	errors := make(chan error, writers)
	for writerID := 0; writerID < writers; writerID++ {
		group.Add(1)
		go func(id int) {
			defer group.Done()
			records := make([]storage.Record, recordsPerWriter)
			for index := range records {
				records[index].EventID = eventID(id, index)
			}
			errors <- store.AppendBatch(context.Background(), records)
		}(writerID)
	}
	group.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatal(err)
		}
	}
	if store.activeBytes == 0 {
		t.Fatal("concurrent appends wrote no bytes")
	}
}

func TestCloseIsIdempotentAndRejectsLaterAppend(t *testing.T) {
	store, err := NewWithClock(testConfig(t.TempDir()), fixedClock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	if err := store.AppendBatch(context.Background(), []storage.Record{{EventID: "closed"}}); err != storage.ErrClosed {
		t.Fatalf("AppendBatch() error = %v, want storage.ErrClosed", err)
	}
}

func bytesLine(data []byte, index int) []byte {
	start := 0
	for current := 0; current < index; current++ {
		for data[start] != '\n' {
			start++
		}
		start++
	}
	end := start
	for end < len(data) && data[end] != '\n' {
		end++
	}
	return data[start:end]
}

func eventID(writerID, index int) string {
	return "writer-" + strconv.Itoa(writerID) + "-" + strconv.Itoa(index)
}
