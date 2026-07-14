package servicelog

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func baseTestFilePolicy(dir string, maxBytes, maxAge int64) FilePolicy {
	return FilePolicy{
		Directory:         dir,
		Prefix:            "game-server",
		SegmentMaxBytes:   maxBytes,
		SegmentMaxAge:     time.Duration(maxAge),
		RetentionMaxAge:   3650 * 24 * time.Hour,
		RetentionMaxBytes: 1 << 60,
	}
}

func testRollingWriter(t *testing.T, dir string, maxBytes, maxAge int64, clock *fakeClock) *rollingWriter {
	t.Helper()
	return testRollingWriterWithPolicy(t, baseTestFilePolicy(dir, maxBytes, maxAge), clock, nil)
}

func testRollingWriterWithPolicy(t *testing.T, policy FilePolicy, clock *fakeClock, mutate func(*runtimeDependencies)) *rollingWriter {
	t.Helper()
	deps := defaultRuntimeDependencies()
	deps.now = clock.now
	if mutate != nil {
		mutate(&deps)
	}
	writer, err := newRollingWriter(policy, deps)
	if err != nil {
		t.Fatalf("newRollingWriter() error = %v", err)
	}
	return writer
}

func writeStaleActiveFile(t *testing.T, dir string, contents []byte, modTime time.Time) string {
	t.Helper()
	activePath := filepath.Join(dir, "active", "game-server.jsonl.open")
	if err := os.MkdirAll(filepath.Dir(activePath), 0o755); err != nil {
		t.Fatalf("MkdirAll(active) error = %v", err)
	}
	if err := os.WriteFile(activePath, contents, 0o644); err != nil {
		t.Fatalf("WriteFile(active) error = %v", err)
	}
	if err := os.Chtimes(activePath, modTime, modTime); err != nil {
		t.Fatalf("Chtimes(active) error = %v", err)
	}
	return activePath
}

func writeArchiveSegmentFile(t *testing.T, path string, contents []byte, modTime time.Time) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s) error = %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, contents, 0o644); err != nil {
		t.Fatalf("WriteFile(%s) error = %v", path, err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("Chtimes(%s) error = %v", path, err)
	}
}

func readGzipFile(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("NewReader(%s) error = %v", path, err)
	}
	defer reader.Close()
	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("ReadAll(%s) error = %v", path, err)
	}
	return decoded
}

func TestRollingWriterRotatesWhenWriteWouldExceedThreshold(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 123456789, time.UTC)
	rotatedAt := startedAt.Add(2 * time.Second)
	clock := &fakeClock{current: startedAt}
	writer := testRollingWriter(t, dir, 10, int64(time.Hour), clock)

	first := []byte("first\n")
	second := []byte("second\n")
	if _, err := writer.Write(first); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	clock.current = rotatedAt
	if _, err := writer.Write(second); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	archivePath := archiveSegmentPath(dir, "game-server", startedAt, rotatedAt, 1)
	archiveData, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile(archive) error = %v", err)
	}
	if !bytes.Equal(archiveData, first) {
		t.Fatalf("archive contents = %q, want %q", string(archiveData), string(first))
	}

	activePath := filepath.Join(dir, "active", "game-server.jsonl.open")
	activeData, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if !bytes.Equal(activeData, second) {
		t.Fatalf("active contents = %q, want %q", string(activeData), string(second))
	}
	if writer.bytesWritten != int64(len(second)) {
		t.Fatalf("bytesWritten = %d, want %d", writer.bytesWritten, len(second))
	}
}

func TestRollingWriterLeavesOversizedRecordUnsplit(t *testing.T) {
	dir := t.TempDir()
	clock := &fakeClock{current: time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)}
	writer := testRollingWriter(t, dir, 5, int64(time.Hour), clock)

	payload := []byte("oversized-record\n")
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if writer.bytesWritten != int64(len(payload)) {
		t.Fatalf("bytesWritten = %d, want %d", writer.bytesWritten, len(payload))
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	activePath := filepath.Join(dir, "active", "game-server.jsonl.open")
	activeData, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if !bytes.Equal(activeData, payload) {
		t.Fatalf("active contents = %q, want %q", string(activeData), string(payload))
	}
	if _, err := os.Stat(filepath.Join(dir, "archive")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("archive directory exists unexpectedly: %v", err)
	}
}

func TestRollingWriterArchivesCompletedSegmentsWithDeterministicPlacement(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.January, 2, 3, 4, 5, 123456789, time.UTC)
	rotatedAt := startedAt.Add(4 * time.Second)
	clock := &fakeClock{current: startedAt}
	writer := testRollingWriter(t, dir, 5, int64(time.Hour), clock)

	first := []byte("one\n")
	second := []byte("two\n")
	if _, err := writer.Write(first); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	clock.current = rotatedAt
	if _, err := writer.Write(second); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	expectedArchive := archiveSegmentPath(dir, "game-server", startedAt, rotatedAt, 1)
	if _, err := os.Stat(expectedArchive); err != nil {
		t.Fatalf("Stat(expected archive) error = %v", err)
	}
	archiveDir := archiveDirectoryForTime(dir, rotatedAt)
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("ReadDir(archive) error = %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("archive entry count = %d, want 1", len(entries))
	}
	if entries[0].Name() != filepath.Base(expectedArchive) {
		t.Fatalf("archive file = %q, want %q", entries[0].Name(), filepath.Base(expectedArchive))
	}
}

func TestRollingWriterMultipleRotationsUseCollisionSafeSequence(t *testing.T) {
	dir := t.TempDir()
	stamp := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	clock := &fakeClock{current: stamp}
	writer := testRollingWriter(t, dir, 4, int64(time.Hour), clock)

	payloads := [][]byte{[]byte("one\n"), []byte("two\n"), []byte("tri\n")}
	for i, payload := range payloads {
		if _, err := writer.Write(payload); err != nil {
			t.Fatalf("Write %d error = %v", i+1, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	archiveDir := archiveDirectoryForTime(dir, stamp)
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		t.Fatalf("ReadDir(archive) error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("archive entry count = %d, want 2", len(entries))
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	expectedFirst := archiveSegmentName("game-server", stamp, stamp, 1)
	expectedSecond := archiveSegmentName("game-server", stamp, stamp, 2)
	if names[0] != expectedFirst || names[1] != expectedSecond {
		t.Fatalf("archive names = %v, want [%s %s]", names, expectedFirst, expectedSecond)
	}

	activePath := filepath.Join(dir, "active", "game-server.jsonl.open")
	activeData, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if !bytes.Equal(activeData, payloads[2]) {
		t.Fatalf("active contents = %q, want %q", string(activeData), string(payloads[2]))
	}
}

func TestRollingWriterDoesNotRotateBeforeAgeThreshold(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	clock := &fakeClock{current: startedAt}
	writer := testRollingWriter(t, dir, 1024, int64(time.Second), clock)

	first := []byte("first\n")
	second := []byte("second\n")
	if _, err := writer.Write(first); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	clock.current = startedAt.Add(999 * time.Millisecond)
	if _, err := writer.Write(second); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	activePath := filepath.Join(dir, "active", "game-server.jsonl.open")
	activeData, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	want := append(append([]byte(nil), first...), second...)
	if !bytes.Equal(activeData, want) {
		t.Fatalf("active contents = %q, want %q", string(activeData), string(want))
	}
	if _, err := os.Stat(filepath.Join(dir, "archive")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("archive directory exists unexpectedly: %v", err)
	}
}

func TestRollingWriterRotatesAtAgeThreshold(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	clock := &fakeClock{current: startedAt}
	writer := testRollingWriter(t, dir, 1024, int64(time.Second), clock)

	first := []byte("first\n")
	second := []byte("second\n")
	if _, err := writer.Write(first); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	clock.current = startedAt.Add(time.Second)
	if _, err := writer.Write(second); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	archivePath := archiveSegmentPath(dir, "game-server", startedAt, startedAt.Add(time.Second), 1)
	archiveData, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile(archive) error = %v", err)
	}
	if !bytes.Equal(archiveData, first) {
		t.Fatalf("archive contents = %q, want %q", string(archiveData), string(first))
	}

	activePath := filepath.Join(dir, "active", "game-server.jsonl.open")
	activeData, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if !bytes.Equal(activeData, second) {
		t.Fatalf("active contents = %q, want %q", string(activeData), string(second))
	}
}

func TestRollingWriterIgnoresAgeThresholdOnEmptySegment(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	clock := &fakeClock{current: startedAt.Add(5 * time.Second)}
	writer := testRollingWriter(t, dir, 1024, int64(time.Second), clock)

	payload := []byte("first\n")
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	activePath := filepath.Join(dir, "active", "game-server.jsonl.open")
	activeData, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if !bytes.Equal(activeData, payload) {
		t.Fatalf("active contents = %q, want %q", string(activeData), string(payload))
	}
	if _, err := os.Stat(filepath.Join(dir, "archive")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("archive directory exists unexpectedly: %v", err)
	}
}

func TestRollingWriterCompressesRotatedSegmentsWhenEnabled(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 123456789, time.UTC)
	rotatedAt := startedAt.Add(2 * time.Second)
	clock := &fakeClock{current: startedAt}
	policy := baseTestFilePolicy(dir, 10, int64(time.Hour))
	policy.CompressionEnabled = true
	writer := testRollingWriterWithPolicy(t, policy, clock, nil)

	first := []byte("first\n")
	second := []byte("second\n")
	if _, err := writer.Write(first); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	clock.current = rotatedAt
	if _, err := writer.Write(second); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	archivePath := compressedArchiveSegmentPath(dir, "game-server", startedAt, rotatedAt, 1)
	archiveData := readGzipFile(t, archivePath)
	if !bytes.Equal(archiveData, first) {
		t.Fatalf("compressed archive contents = %q, want %q", string(archiveData), string(first))
	}

	activePath := filepath.Join(dir, "active", "game-server.jsonl.open")
	activeData, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if !bytes.Equal(activeData, second) {
		t.Fatalf("active contents = %q, want %q", string(activeData), string(second))
	}
}

func TestRollingWriterLeavesUncompressedArchiveWhenCompressionFails(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	rotatedAt := startedAt.Add(2 * time.Second)
	clock := &fakeClock{current: startedAt}
	policy := baseTestFilePolicy(dir, 10, int64(time.Hour))
	policy.CompressionEnabled = true
	writer := testRollingWriterWithPolicy(t, policy, clock, func(deps *runtimeDependencies) {
		baseOpen := deps.openFile
		deps.openFile = func(path string, flag int, perm fs.FileMode) (*os.File, error) {
			if strings.HasSuffix(path, ".tmp") {
				return nil, errors.New("compression open failure")
			}
			return baseOpen(path, flag, perm)
		}
	})

	first := []byte("first\n")
	second := []byte("second\n")
	if _, err := writer.Write(first); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	clock.current = rotatedAt
	if _, err := writer.Write(second); err == nil {
		t.Fatalf("second Write() error = nil, want compression failure")
	}

	archivePath := archiveSegmentPath(dir, "game-server", startedAt, rotatedAt, 1)
	archiveData, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile(uncompressed archive) error = %v", err)
	}
	if !bytes.Equal(archiveData, first) {
		t.Fatalf("archive contents = %q, want %q", string(archiveData), string(first))
	}
	if _, err := os.Stat(compressedArchiveSegmentPath(dir, "game-server", startedAt, rotatedAt, 1)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("compressed archive exists unexpectedly: %v", err)
	}
}

func TestRollingWriterRecoversCompleteActiveSegment(t *testing.T) {
	dir := t.TempDir()
	staleModTime := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	now := staleModTime.Add(5 * time.Second)
	stalePath := writeStaleActiveFile(t, dir, []byte("one\ntwo\n"), staleModTime)
	clock := &fakeClock{current: now}
	writer := testRollingWriter(t, dir, 1024, int64(time.Hour), clock)
	defer writer.Close()

	archivePath := archiveSegmentPath(dir, "game-server", staleModTime, now, 1)
	archiveData, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile(archive) error = %v", err)
	}
	if !bytes.Equal(archiveData, []byte("one\ntwo\n")) {
		t.Fatalf("archive contents = %q, want %q", string(archiveData), "one\ntwo\n")
	}

	activeData, err := os.ReadFile(stalePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if len(activeData) != 0 {
		t.Fatalf("active contents = %q, want empty fresh file", string(activeData))
	}
}

func TestRollingWriterRecoversTruncatedFinalLine(t *testing.T) {
	dir := t.TempDir()
	staleModTime := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	now := staleModTime.Add(5 * time.Second)
	stalePath := writeStaleActiveFile(t, dir, []byte("one\npartial"), staleModTime)
	clock := &fakeClock{current: now}
	writer := testRollingWriter(t, dir, 1024, int64(time.Hour), clock)
	defer writer.Close()

	archivePath := archiveSegmentPath(dir, "game-server", staleModTime, now, 1)
	archiveData, err := os.ReadFile(archivePath)
	if err != nil {
		t.Fatalf("ReadFile(archive) error = %v", err)
	}
	if !bytes.Equal(archiveData, []byte("one\n")) {
		t.Fatalf("archive contents = %q, want %q", string(archiveData), "one\n")
	}

	activeData, err := os.ReadFile(stalePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if len(activeData) != 0 {
		t.Fatalf("active contents = %q, want empty fresh file", string(activeData))
	}
}

func TestRollingWriterRemovesEmptyStaleActiveFile(t *testing.T) {
	dir := t.TempDir()
	staleModTime := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	stalePath := writeStaleActiveFile(t, dir, nil, staleModTime)
	clock := &fakeClock{current: staleModTime.Add(5 * time.Second)}
	writer := testRollingWriter(t, dir, 1024, int64(time.Hour), clock)
	defer writer.Close()

	activeData, err := os.ReadFile(stalePath)
	if err != nil {
		t.Fatalf("ReadFile(active) error = %v", err)
	}
	if len(activeData) != 0 {
		t.Fatalf("active contents = %q, want empty fresh file", string(activeData))
	}
	if _, err := os.Stat(filepath.Join(dir, "archive")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("archive directory exists unexpectedly: %v", err)
	}
}

func TestRollingWriterRunsRetentionAfterSuccessfulRotation(t *testing.T) {
	dir := t.TempDir()
	startedAt := time.Date(2026, time.July, 14, 15, 4, 5, 0, time.UTC)
	staleModTime := startedAt.Add(-50 * time.Minute)
	stalePath := archiveSegmentPath(dir, "game-server", staleModTime, staleModTime, 1)
	writeArchiveSegmentFile(t, stalePath, []byte("stale\n"), staleModTime)

	clock := &fakeClock{current: startedAt}
	policy := FilePolicy{
		Directory:         dir,
		Prefix:            "game-server",
		SegmentMaxBytes:   10,
		SegmentMaxAge:     time.Hour,
		RetentionMaxAge:   time.Hour,
		RetentionMaxBytes: 1024,
	}
	writer := testRollingWriterWithPolicy(t, policy, clock, nil)
	defer writer.Close()

	first := []byte("first\n")
	second := []byte("second\n")
	if _, err := writer.Write(first); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	clock.current = startedAt.Add(20 * time.Minute)
	if _, err := writer.Write(second); err != nil {
		t.Fatalf("second Write() error = %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Fatalf("stale archive still exists: %v", err)
	}
}
