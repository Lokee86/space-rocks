package servicelog

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

type fakeClock struct {
	current time.Time
}

func (c *fakeClock) now() time.Time {
	return c.current
}

func (c *fakeClock) advance(delta time.Duration) {
	c.current = c.current.Add(delta)
}

type fakeTicker struct {
	ch        chan time.Time
	stopCount int
}

func (t *fakeTicker) C() <-chan time.Time {
	return t.ch
}

func (t *fakeTicker) Stop() {
	t.stopCount++
}

type fakeFilesystem struct {
	mkdir    func(string, fs.FileMode) error
	openFile func(string, int, fs.FileMode) (io.WriteCloser, error)
	readDir  func(string) ([]os.DirEntry, error)
	rename   func(string, string) error
	remove   func(string) error
	stat     func(string) (os.FileInfo, error)
}

func (f fakeFilesystem) dependencies(writer io.Writer, clock *fakeClock) runtimeDependencies {
	return runtimeDependencies{
		consoleWriter: writer,
		now:           clock.now,
		mkdir:         f.mkdir,
		openFile:      f.openFile,
		readFile:      os.ReadFile,
		readDir:       f.readDir,
		rename:        f.rename,
		remove:        f.remove,
		stat:          f.stat,
	}
}

func testRuntimeDependencies(clock *fakeClock) runtimeDependencies {
	return runtimeDependencies{
		consoleWriter: io.Discard,
		now:           clock.now,
		mkdir:         os.MkdirAll,
		openFile: func(name string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			return os.OpenFile(name, flag, perm)
		},
		readFile: os.ReadFile,
		readDir:  os.ReadDir,
		rename:   os.Rename,
		remove:   os.Remove,
		stat:     os.Stat,
	}
}

func readJSONLines(t *testing.T, path string) []map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	return readJSONLinesFromBytes(t, data)
}

func readGzipJSONLines(t *testing.T, path string) []map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("gzip.NewReader(%q) error = %v", path, err)
	}
	defer reader.Close()

	decoded, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("io.ReadAll(%q) error = %v", path, err)
	}
	return readJSONLinesFromBytes(t, decoded)
}

func readJSONLinesFromBytes(t *testing.T, data []byte) []map[string]any {
	t.Helper()

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	records := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatalf("json.Unmarshal() error = %v; line = %q", err, string(line))
		}
		records = append(records, record)
	}
	return records
}

func collectLogFiles(t *testing.T, directory string) []string {
	t.Helper()

	var paths []string
	if err := filepath.WalkDir(directory, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".jsonl") || strings.HasSuffix(path, ".jsonl.open") || strings.HasSuffix(path, ".jsonl.gz") {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("filepath.WalkDir(%q) error = %v", directory, err)
	}
	sort.Strings(paths)
	return paths
}

func assertArchiveFilename(t *testing.T, path string, prefix string) {
	t.Helper()

	name := filepath.Base(path)
	pattern := regexp.MustCompile("^" + regexp.QuoteMeta(prefix) + `-\d{8}T\d{6}\.\d{9}Z-\d{8}T\d{6}\.\d{9}Z(?:-\d+)?\.jsonl$`)
	if !pattern.MatchString(name) {
		t.Fatalf("archive filename %q does not match expected pattern", name)
	}
}
