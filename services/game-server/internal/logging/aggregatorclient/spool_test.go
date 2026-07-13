package aggregatorclient

import (
	"os"
	"path/filepath"
	"testing"
)

func testSpool(t *testing.T, cap int64) *spoolStore {
	t.Helper()
	return &spoolStore{directory: t.TempDir(), byteCap: cap}
}

func TestSpoolSaveAndLoadIsAtomic(t *testing.T) {
	spool := testSpool(t, 1024)
	batch := []byte(`{"events":[{"id":1}]}`)
	if _, err := spool.save(batch); err != nil {
		t.Fatal(err)
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 1 {
		t.Fatalf("pending count = %d, want 1", len(pending))
	}
	loaded, err := os.ReadFile(pending[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(loaded) != string(batch) {
		t.Fatalf("loaded = %q, want %q", loaded, batch)
	}
	if _, err := os.Stat(filepath.Join(spool.directory, ".aggregator-batch-")); !os.IsNotExist(err) {
		t.Fatal("temporary spool file remains")
	}
}

func TestSpoolReturnsOldestFirst(t *testing.T) {
	spool := testSpool(t, 1024)
	for _, batch := range [][]byte{[]byte("one"), []byte("two"), []byte("three")} {
		if _, err := spool.save(batch); err != nil {
			t.Fatal(err)
		}
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 3 {
		t.Fatalf("pending count = %d, want 3", len(pending))
	}
	for index, want := range []string{"one", "two", "three"} {
		got, err := os.ReadFile(pending[index].Path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Fatalf("batch %d = %q, want %q", index, got, want)
		}
	}
}

func TestSpoolRemoval(t *testing.T) {
	spool := testSpool(t, 1024)
	if _, err := spool.save([]byte("remove me")); err != nil {
		t.Fatal(err)
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if err := spool.remove(pending[0]); err != nil {
		t.Fatal(err)
	}
	pending, err = spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending count = %d, want 0", len(pending))
	}
}

func TestSpoolEvictsOldestFilesAtCap(t *testing.T) {
	spool := testSpool(t, 1<<20)
	first, _ := encodeBatch([][]byte{[]byte(`{"id":1}`)})
	second, _ := encodeBatch([][]byte{[]byte(`{"id":2}`)})
	third, _ := encodeBatch([][]byte{[]byte(`{"id":3}`)})
	spool.byteCap = int64(len(first) + len(second))
	if _, err := spool.save(first); err != nil {
		t.Fatal(err)
	}
	if _, err := spool.save(second); err != nil {
		t.Fatal(err)
	}
	result, err := spool.save(third)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Stored || result.EvictedBatches != 1 || result.EvictedEvents != 1 {
		t.Fatalf("save result = %#v", result)
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 {
		t.Fatalf("pending count = %d, want 2", len(pending))
	}
	got, err := os.ReadFile(pending[1].Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(third) {
		t.Fatal("oldest batch was not evicted first")
	}
}

func TestSpoolRejectsOversizedBatch(t *testing.T) {
	spool := testSpool(t, 4)
	if _, err := spool.save([]byte("12345")); err == nil {
		t.Fatal("expected oversized batch error")
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending count = %d, want 0", len(pending))
	}
}

func TestSpoolLoadValidatesAndCountsEvents(t *testing.T) {
	spool := testSpool(t, 1024)
	payload, _ := encodeBatch([][]byte{[]byte(`{"id":1}`), []byte(`{"id":2}`)})
	if _, err := spool.save(payload); err != nil {
		t.Fatal(err)
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	loaded, events, err := spool.load(pending[0])
	if err != nil || string(loaded) != string(payload) || events != 2 {
		t.Fatalf("load = %q, %d, %v", loaded, events, err)
	}
}

func TestSpoolLoadRejectsMalformedBatch(t *testing.T) {
	spool := testSpool(t, 1024)
	path := filepath.Join(spool.directory, spoolFilePrefix+"malformed"+spoolFileSuffix)
	if err := os.WriteFile(path, []byte("not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := spool.load(pending[0]); err == nil {
		t.Fatal("expected malformed batch error")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestSpoolLoadRejectsOversizedPendingFile(t *testing.T) {
	spool := testSpool(t, 4)
	path := filepath.Join(spool.directory, spoolFilePrefix+"oversized"+spoolFileSuffix)
	if err := os.WriteFile(path, []byte("12345"), 0o644); err != nil {
		t.Fatal(err)
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := spool.load(pending[0]); err == nil {
		t.Fatal("expected oversized pending batch error")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestSpoolIgnoresUnrelatedFiles(t *testing.T) {
	spool := testSpool(t, 1024)
	if err := os.WriteFile(filepath.Join(spool.directory, "notes.txt"), []byte("ignore"), 0o644); err != nil {
		t.Fatal(err)
	}
	pending, err := spool.pending()
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 0 {
		t.Fatalf("pending count = %d, want 0", len(pending))
	}
}
