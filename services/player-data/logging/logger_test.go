package logging

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/player-data/observability"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

func testIdentity() servicelog.ServiceIdentity {
	return servicelog.ServiceIdentity{
		Name:        ServiceName,
		Version:     "test-build",
		Environment: "test",
		InstanceID:  "550e8400-e29b-41d4-a716-446655440001",
	}
}

func TestCategoryFilteringHonorsCategoryLevelEnvironment(t *testing.T) {
	restore := captureStderr(t)
	t.Cleanup(func() { _ = CloseFileOutput(); Configure("warn") })
	t.Setenv(EnvHTTPLevel, "error")
	Configure("info")
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}

	HTTP.Info("http filtered")
	Store.Info("store emitted")
	output := restore()
	if strings.Contains(output, "http filtered") || !strings.Contains(output, "store emitted") {
		t.Fatalf("output=%q", output)
	}
}

func TestFileOutputUsesPlayerDataIdentityAndGeneratedPolicy(t *testing.T) {
	restore := captureStderr(t)
	defer restore()
	t.Cleanup(func() { _ = CloseFileOutput() })
	Configure("info")
	id := testIdentity()
	if err := ConfigureRuntime(id); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	path, err := ConfigureFileOutput(dir, "player-data")
	if err != nil {
		t.Fatal(err)
	}
	Runtime.Info("player runtime configured")
	if err := CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, "player-data.jsonl.open") {
		t.Fatalf("path=%q", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, value := range []string{`"service":"player-data"`, `"build_version":"test-build"`, `"environment":"test"`, `"service_instance_id":"550e8400-e29b-41d4-a716-446655440001"`} {
		if !strings.Contains(content, value) {
			t.Fatalf("missing %s in %s", value, content)
		}
	}
	cfg := fileRuntimeConfig(id, dir, "player-data")
	if cfg.File.SegmentMaxAge != time.Second*time.Duration(observability.FileLoggingMaxActiveSegmentAgeSeconds) || cfg.File.RetentionMaxAge != time.Second*time.Duration(observability.RetentionDefaultAgeOperationalSeconds) || cfg.File.CompressionEnabled != observability.FileLoggingCompressionEnabled {
		t.Fatalf("policy=%#v", cfg.File)
	}
}

func TestFileOutputRecoversAndCompressesInterruptedSegment(t *testing.T) {
	directory := t.TempDir()
	activePath := filepath.Join(directory, "player-data.jsonl.open")
	if err := os.WriteFile(activePath, []byte("{\"message\":\"interrupted\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = CloseFileOutput() })
	if _, err := ConfigureFileOutput(directory, "player-data"); err != nil {
		t.Fatal(err)
	}
	archives, err := filepath.Glob(filepath.Join(directory, "archive", "*", "*.jsonl.gz"))
	if err != nil || len(archives) != 1 {
		t.Fatalf("archives=%v err=%v", archives, err)
	}
}
func TestFileFailureDegradesStatusAndLeavesConsoleOperational(t *testing.T) {
	restore := captureStderr(t)
	t.Cleanup(func() { _ = CloseFileOutput() })
	Configure("info")
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ConfigureFileOutput(filepath.Join(blocked, "logs"), "player-data"); err != nil {
		t.Fatal(err)
	}
	Server.Info("console survives")
	output := restore()
	if !Status().Degraded || !strings.Contains(output, "console survives") {
		t.Fatalf("status=%#v output=%q", Status(), output)
	}
}

func TestCloseFileOutputIsIdempotent(t *testing.T) {
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}
	if _, err := ConfigureFileOutput(t.TempDir(), "player-data"); err != nil {
		t.Fatal(err)
	}
	if err := CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if err := CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if !Status().Closed {
		t.Fatalf("status=%#v", Status())
	}
}

func TestConfigureRuntimeRejectsIncompleteIdentity(t *testing.T) {
	if err := ConfigureRuntime(servicelog.ServiceIdentity{Name: ServiceName}); err == nil {
		t.Fatal("expected incomplete identity error")
	}
}

func captureStderr(t *testing.T) func() string {
	t.Helper()
	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = writer
	closed := false
	t.Cleanup(func() {
		if !closed {
			os.Stderr = original
			_ = writer.Close()
			_ = reader.Close()
		}
	})
	return func() string {
		if closed {
			return ""
		}
		closed = true
		os.Stderr = original
		_ = writer.Close()
		data, _ := io.ReadAll(reader)
		_ = reader.Close()
		return string(data)
	}
}
