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
	closeErr   error
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
		"service":            "game-server",
		"service_instance_id": "instance-1",
		"environment":        "test",
		"build_version":      "dev",
		"msg":                "started",
		"phase":              "init",
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
