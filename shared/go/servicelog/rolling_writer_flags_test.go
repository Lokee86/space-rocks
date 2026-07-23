package servicelog

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRotationReopensActiveFileWithPortableWritableFlags(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	clock := &fakeClock{current: time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)}
	policy := validFileConfig(directory)
	policy.SegmentMaxBytes = 1

	deps := testRuntimeDependencies(clock)
	var openFlags []int
	deps.openFile = func(name string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
		openFlags = append(openFlags, flag)
		return os.OpenFile(name, flag, perm)
	}

	writer, err := newRollingWriter(policy, deps)
	if err != nil {
		t.Fatalf("newRollingWriter() error = %v", err)
	}
	defer writer.Close()

	if _, err := writer.Write([]byte("one\n")); err != nil {
		t.Fatalf("first Write() error = %v", err)
	}
	if _, err := writer.Write([]byte("two\n")); err != nil {
		t.Fatalf("rotating Write() error = %v", err)
	}

	if len(openFlags) < 2 {
		t.Fatalf("open call count = %d, want at least 2", len(openFlags))
	}
	want := os.O_CREATE | os.O_TRUNC | os.O_WRONLY
	if got := openFlags[1]; got != want {
		t.Fatalf("rotated active-file flags = %d, want portable writable flags %d", got, want)
	}
}
