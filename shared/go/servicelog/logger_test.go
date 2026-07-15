package servicelog

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type trackedWriteCloser struct {
	bytes.Buffer
	closeCount int
	syncCount  int
	writeCount int
	closeErr   error
	syncErr    error
	writeErr   error
}

func (w *trackedWriteCloser) Write(p []byte) (int, error) {
	w.writeCount++
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	return w.Buffer.Write(p)
}

func (w *trackedWriteCloser) Sync() error {
	w.syncCount++
	return w.syncErr
}

func (w *trackedWriteCloser) Close() error {
	w.closeCount++
	return w.closeErr
}

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
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}

	lines := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	if len(lines) != 1 {
		t.Fatalf("log file line count = %d, want 1; contents = %q", len(lines), string(data))
	}

	var record map[string]any
	if err := json.Unmarshal(lines[0], &record); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; line = %q", err, string(lines[0]))
	}

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
	var console bytes.Buffer
	sink := &trackedWriteCloser{}

	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig("logs"),
		FileEnabled:    true,
		ConsoleEnabled: true,
	}, runtimeDependencies{
		consoleWriter: &console,
		mkdir: func(string, fs.FileMode) error {
			return nil
		},
		openFile: func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return sink, nil
		},
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
	if sink.Len() == 0 {
		t.Fatal("file output is empty")
	}
	if !strings.Contains(sink.String(), `"msg":"ready"`) {
		t.Fatalf("file output missing JSON record: %q", sink.String())
	}
	if sink.closeCount != 1 {
		t.Fatalf("close count = %d, want 1", sink.closeCount)
	}
}

func TestWriteRecordPreservesCanonicalJSONWithoutSlogReshaping(t *testing.T) {
	var console bytes.Buffer
	sink := &trackedWriteCloser{}
	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig("logs"),
		FileEnabled:    true,
		ConsoleEnabled: true,
	}, runtimeDependencies{
		consoleWriter: &console,
		mkdir:         func(string, fs.FileMode) error { return nil },
		openFile:      func(string, int, fs.FileMode) (io.WriteCloser, error) { return sink, nil },
	})
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer runtime.Close()

	payload := []byte(`{"event":"log_message","service":"game-server","message":"ready"}`)
	if err := runtime.WriteRecord(payload, "[game-server][info] ready"); err != nil {
		t.Fatalf("WriteRecord() error = %v", err)
	}
	if got, want := sink.String(), string(payload)+"\n"; got != want {
		t.Fatalf("file output = %q, want exact canonical payload %q", got, want)
	}
	if got, want := console.String(), "[game-server][info] ready\n"; got != want {
		t.Fatalf("console output = %q, want %q", got, want)
	}
}

func TestRepeatedRotationUsesDeterministicArchiveSuffixForIdenticalTimestamps(t *testing.T) {
	directory := filepath.Join(t.TempDir(), "logs")
	now := time.Date(2026, 7, 14, 13, 30, 0, 0, time.UTC)
	clock := &fakeClock{current: now}

	runtime, err := openWithDependencies(Config{
		Identity: ServiceIdentity{Name: "game-server"},
		File: func() FilePolicy {
			cfg := validFileConfig(directory)
			cfg.SegmentMaxBytes = 1
			cfg.SegmentMaxAge = time.Hour
			return cfg
		}(),
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, runtimeDependencies{
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
	})
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	defer func() { _ = runtime.Close() }()

	runtime.Logger().Info("one")
	runtime.Logger().Info("two")
	runtime.Logger().Info("three")

	paths := collectLogFiles(t, directory)
	if len(paths) != 3 {
		t.Fatalf("log file count = %d, want 3; paths = %v", len(paths), paths)
	}
	activePath := filepath.Join(directory, "game-server.jsonl.open")
	var baseArchiveSeen bool
	var suffixedArchiveSeen bool
	for _, path := range paths {
		if path == activePath {
			continue
		}
		assertArchiveFilename(t, path, "game-server")
		base := filepath.Base(path)
		if strings.HasSuffix(base, "-2.jsonl") {
			suffixedArchiveSeen = true
		}
		if strings.Contains(base, "-20260714T133000.000000000Z-20260714T133000.000000000Z.jsonl") && !strings.Contains(base, "-2.jsonl") {
			baseArchiveSeen = true
		}
	}
	if !baseArchiveSeen {
		t.Fatal("base archive name not found")
	}
	if !suffixedArchiveSeen {
		t.Fatal("collision-safe suffixed archive name not found")
	}
}

func TestOpenFallsBackToActiveFileOpenFailure(t *testing.T) {
	wantErr := errors.New("open failed")

	runtime, err := openWithDependencies(Config{
		Identity:       ServiceIdentity{Name: "game-server"},
		File:           validFileConfig("logs"),
		FileEnabled:    true,
		ConsoleEnabled: false,
	}, runtimeDependencies{
		mkdir: func(string, fs.FileMode) error { return nil },
		openFile: func(string, int, fs.FileMode) (io.WriteCloser, error) {
			return nil, wantErr
		},
	})
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v, want nil", err)
	}
	if runtime == nil {
		t.Fatal("openWithDependencies() returned nil runtime")
	}
	status := runtime.Status()
	if !status.Degraded {
		t.Fatal("runtime is not degraded after open failure")
	}
	if status.FailureCount != 1 {
		t.Fatalf("failure count = %d, want 1", status.FailureCount)
	}
	if status.LastError != wantErr.Error() {
		t.Fatalf("last error = %q, want %q", status.LastError, wantErr.Error())
	}
}

func TestRuntimeCloseClosesActiveFileOnce(t *testing.T) {
	closeErr := errors.New("close failed")
	sink := &trackedWriteCloser{closeErr: closeErr}

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
			return sink, nil
		},
	})
	if err != nil {
		t.Fatalf("openWithDependencies() error = %v", err)
	}
	if runtime.Status().Closed {
		t.Fatal("runtime starts closed")
	}

	if err := runtime.Close(); !errors.Is(err, closeErr) {
		t.Fatalf("first Close() error = %v, want %v", err, closeErr)
	}
	status := runtime.Status()
	if !status.Degraded {
		t.Fatal("runtime is not degraded after Close() failure")
	}
	if status.FailureCount != 1 {
		t.Fatalf("failure count after Close() = %d, want 1", status.FailureCount)
	}
	if status.LastError != closeErr.Error() {
		t.Fatalf("last error after Close() = %q, want %q", status.LastError, closeErr.Error())
	}
	if sink.closeCount != 1 {
		t.Fatalf("close count after first Close() = %d, want 1", sink.closeCount)
	}
	if err := runtime.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	if sink.closeCount != 1 {
		t.Fatalf("close count after second Close() = %d, want 1", sink.closeCount)
	}
	if !runtime.Status().Closed {
		t.Fatal("runtime remains open after Close()")
	}
}
