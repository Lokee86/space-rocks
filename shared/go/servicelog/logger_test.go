package servicelog

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func validFileConfig(directory string) FilePolicy {
	return FilePolicy{
		Directory:         directory,
		Prefix:            "game-server",
		SegmentMaxBytes:   1024,
		SegmentMaxAge:     time.Hour,
		RetentionMaxAge:   24 * time.Hour,
		RetentionMaxBytes: 4096,
	}
}

func openTestRuntime(t *testing.T, directory string, clock *fakeClock, config FilePolicy, identity ServiceIdentity) (*Runtime, *trackedFile) {
	t.Helper()

	var current *trackedFile
	runtime, err := openWithDependencies(Config{
		Identity:       identity,
		File:           config,
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, runtimeDependencies{
		consoleWriter: io.Discard,
		now:           clock.now,
		mkdir: func(path string, perm fs.FileMode) error {
			return os.MkdirAll(path, perm)
		},
		openFile: func(name string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			file, err := os.OpenFile(name, flag, perm)
			if err != nil {
				return nil, err
			}
			current = &trackedFile{File: file}
			return current, nil
		},
		readDir: os.ReadDir,
		rename:  os.Rename,
		remove:  os.Remove,
		stat:    os.Stat,
	})
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	if current == nil {
		t.Fatal("openFile was not called")
	}
	return runtime, current
}

func readJSONLRecords(t *testing.T, path string) []map[string]any {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return nil
	}

	lines := bytes.Split(trimmed, []byte("\n"))
	records := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatalf("json.Unmarshal() error = %v; line = %q", err, string(line))
		}
		records = append(records, record)
	}
	return records
}

func measureJSONLineLen(identity ServiceIdentity, msg string, args ...any) int {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil)).With(
		slog.String("service", identity.Name),
	)
	if identity.InstanceID != "" {
		logger = logger.With(slog.String("service_instance_id", identity.InstanceID))
	}
	if identity.Environment != "" {
		logger = logger.With(slog.String("environment", identity.Environment))
	}
	if identity.Version != "" {
		logger = logger.With(slog.String("build_version", identity.Version))
	}
	logger.Info(msg, args...)
	return buf.Len()
}

func TestOpenCreatesActiveFilePath(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "nested", "logs")
	runtime, err := Open(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig(directory),
		FileEnabled:    true,
		ConsoleEnabled: false,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer runtime.Close()

	path := filepath.Join(directory, "game-server.jsonl.open")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("os.Stat(%q) error = %v", path, err)
	}
}

func TestActiveFileWritesJSONLWithServiceIdentity(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	runtime, err := Open(Config{
		Identity: ServiceIdentity{
			Name:        "game-server",
			InstanceID:  "instance-1",
			Environment: "test",
			Version:     "dev",
		},
		File:           validFileConfig(directory),
		FileEnabled:    true,
		ConsoleEnabled: false,
	})
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	runtime.Logger().Info("started", "phase", "init")
	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	path := filepath.Join(directory, "game-server.jsonl.open")
	records := readJSONLRecords(t, path)
	if len(records) != 1 {
		t.Fatalf("log file line count = %d, want 1", len(records))
	}

	record := records[0]
	for key, want := range map[string]string{
		"service":             "game-server",
		"service_instance_id": "instance-1",
		"environment":         "test",
		"build_version":       "dev",
		"msg":                 "started",
		"phase":               "init",
	} {
		got, ok := record[key]
		if !ok {
			t.Fatalf("JSON record missing %q: %v", key, record)
		}
		if got != want {
			t.Fatalf("JSON record %q = %v, want %q", key, got, want)
		}
	}
	if got, ok := record["level"].(string); !ok || got != "INFO" {
		t.Fatalf("JSON record level = %v, want INFO", record["level"])
	}
}

func TestOpenFansOutToConsoleAndFile(t *testing.T) {
	directory := t.TempDir()
	var console bytes.Buffer
	clock := &fakeClock{current: time.Date(2026, time.July, 14, 15, 16, 17, 0, time.UTC)}
	var current *trackedFile
	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig(directory),
		FileEnabled:    true,
		ConsoleEnabled: true,
	}, runtimeDependencies{
		consoleWriter: &console,
		now:           clock.now,
		mkdir: func(path string, perm fs.FileMode) error {
			return os.MkdirAll(path, perm)
		},
		openFile: func(name string, flag int, perm fs.FileMode) (io.WriteCloser, error) {
			file, err := os.OpenFile(name, flag, perm)
			if err != nil {
				return nil, err
			}
			current = &trackedFile{File: file}
			return current, nil
		},
		readDir: os.ReadDir,
		rename:  os.Rename,
		remove:  os.Remove,
		stat:    os.Stat,
	})
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}

	runtime.Logger().Info("ready")
	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	if console.Len() == 0 {
		t.Fatal("console output is empty")
	}
	if !strings.Contains(console.String(), "ready") {
		t.Fatalf("console output missing record: %q", console.String())
	}
	if current == nil {
		t.Fatal("file writer was not opened")
	}
	if current.closeCount != 1 {
		t.Fatalf("file close count = %d, want 1", current.closeCount)
	}

	path := filepath.Join(directory, "game-server.jsonl.open")
	records := readJSONLRecords(t, path)
	if len(records) != 1 {
		t.Fatalf("file output line count = %d, want 1", len(records))
	}
	if got := records[0]["msg"]; got != "ready" {
		t.Fatalf("file output msg = %v, want ready", got)
	}
}

func TestRollingWriterRotatesOnSizeAndContinuesWrites(t *testing.T) {
	directory := t.TempDir()
	clock := &fakeClock{current: time.Date(2026, time.July, 14, 15, 16, 17, 0, time.UTC)}
	identity := ServiceIdentity{Name: "game-server"}
	firstMsg := strings.Repeat("a", 240)
	secondMsg := "second"
	thirdMsg := "third"
	firstLen := measureJSONLineLen(identity, firstMsg, "phase", "first")
	secondLen := measureJSONLineLen(identity, secondMsg, "phase", "second")
	thirdLen := measureJSONLineLen(identity, thirdMsg, "phase", "third")
	if firstLen <= secondLen+thirdLen {
		t.Fatalf("first record size = %d, need it to exceed the combined short record sizes %d", firstLen, secondLen+thirdLen)
	}

	config := validFileConfig(directory)
	config.SegmentMaxBytes = int64(secondLen + thirdLen + 1)
	runtime, _ := openTestRuntime(t, directory, clock, config, identity)

	runtime.Logger().Info(firstMsg, "phase", "first")
	runtime.Logger().Info(secondMsg, "phase", "second")
	runtime.Logger().Info(thirdMsg, "phase", "third")
	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("os.ReadDir(%q) error = %v", directory, err)
	}
	var archiveName string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".open") {
			continue
		}
		archiveName = name
	}
	if archiveName == "" {
		t.Fatal("archive file was not created")
	}
	archivePattern := regexp.MustCompile(`^game-server\.\d{8}T\d{6}\.\d{9}Z-\d{8}T\d{6}\.\d{9}Z\.jsonl$`)
	if !archivePattern.MatchString(archiveName) {
		t.Fatalf("archive name %q does not match expected timestamp pattern", archiveName)
	}

	archiveRecords := readJSONLRecords(t, filepath.Join(directory, archiveName))
	if len(archiveRecords) != 1 {
		t.Fatalf("archive record count = %d, want 1", len(archiveRecords))
	}
	if got := archiveRecords[0]["msg"]; got != firstMsg {
		t.Fatalf("archive msg = %v, want %q", got, firstMsg)
	}

	activeRecords := readJSONLRecords(t, filepath.Join(directory, "game-server.jsonl.open"))
	if len(activeRecords) != 2 {
		t.Fatalf("active record count = %d, want 2", len(activeRecords))
	}
	if got := activeRecords[0]["msg"]; got != secondMsg {
		t.Fatalf("active first msg = %v, want %q", got, secondMsg)
	}
	if got := activeRecords[1]["msg"]; got != thirdMsg {
		t.Fatalf("active second msg = %v, want %q", got, thirdMsg)
	}
}

func TestRollingWriterRotatesOnAge(t *testing.T) {
	directory := t.TempDir()
	clock := &fakeClock{current: time.Date(2026, time.July, 14, 15, 16, 17, 0, time.UTC)}
	identity := ServiceIdentity{Name: "game-server"}
	config := validFileConfig(directory)
	config.SegmentMaxBytes = 4096
	config.SegmentMaxAge = time.Minute
	runtime, _ := openTestRuntime(t, directory, clock, config, identity)

	runtime.Logger().Info("first", "phase", "one")
	clock.advance(time.Minute + time.Second)
	runtime.Logger().Info("second", "phase", "two")
	if err := runtime.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatalf("os.ReadDir(%q) error = %v", directory, err)
	}
	var archiveName string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".open") {
			continue
		}
		archiveName = name
	}
	if archiveName == "" {
		t.Fatal("archive file was not created")
	}

	archiveRecords := readJSONLRecords(t, filepath.Join(directory, archiveName))
	if len(archiveRecords) != 1 {
		t.Fatalf("archive record count = %d, want 1", len(archiveRecords))
	}
	if got := archiveRecords[0]["msg"]; got != "first" {
		t.Fatalf("archive msg = %v, want first", got)
	}

	activeRecords := readJSONLRecords(t, filepath.Join(directory, "game-server.jsonl.open"))
	if len(activeRecords) != 1 {
		t.Fatalf("active record count = %d, want 1", len(activeRecords))
	}
	if got := activeRecords[0]["msg"]; got != "second" {
		t.Fatalf("active msg = %v, want second", got)
	}
}

func TestOpenPropagatesActiveFileOpenFailure(t *testing.T) {
	wantErr := errors.New("open failed")

	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig("logs"),
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, runtimeDependencies{
		mkdir: func(string, fs.FileMode) error {
			return nil
		},
		openFile: func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return nil, wantErr
		},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("openWithDependencies() error = %v, want %v", err, wantErr)
	}
	if runtime != nil {
		t.Fatal("openWithDependencies() returned runtime for open failure")
	}
}

func TestRuntimeCloseClosesRollingWriterOnce(t *testing.T) {
	directory := t.TempDir()
	clock := &fakeClock{current: time.Date(2026, time.July, 14, 15, 16, 17, 0, time.UTC)}
	runtime, current := openTestRuntime(t, directory, clock, validFileConfig(directory), ServiceIdentity{Name: "game-server"})

	runtime.Logger().Info("ready")
	if err := runtime.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if current.closeCount != 1 {
		t.Fatalf("close count after first Close() = %d, want 1", current.closeCount)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if current.closeCount != 1 {
		t.Fatalf("close count after second Close() = %d, want 1", current.closeCount)
	}
	if !runtime.Status().Closed {
		t.Fatal("runtime remains open after Close()")
	}
}
