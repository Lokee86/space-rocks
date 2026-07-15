package logging

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

func testIdentity() servicelog.ServiceIdentity {
	return servicelog.ServiceIdentity{
		Name:        ServiceName,
		Version:     "test-build",
		Environment: "test",
		InstanceID:  "550e8400-e29b-41d4-a716-446655440000",
	}
}

func TestCategoryFilteringHonorsCategoryLevelEnvironment(t *testing.T) {
	resetLogging(t)
	restore := captureStderr(t)
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvGameLevel, "error")
	Configure("info")

	Game.Info("game filtered")
	Network.Info("network emitted")
	output := restore()
	if strings.Contains(output, "game filtered") {
		t.Fatalf("expected game category info to be filtered, got %q", output)
	}
	if !strings.Contains(output, "network emitted") {
		t.Fatalf("expected network category info to be emitted, got %q", output)
	}
}

func TestConfigureFileOutputReturnsActivePathAndWritesRecords(t *testing.T) {
	resetLogging(t)
	restore := captureStderr(t)
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}
	Configure("info")

	directory := t.TempDir()
	path, err := ConfigureFileOutput(directory, "game-server")
	if err != nil {
		t.Fatalf("ConfigureFileOutput() error = %v", err)
	}
	wantPath := filepath.Join(directory, "game-server.jsonl.open")
	if path != wantPath {
		t.Fatalf("ConfigureFileOutput() path = %q, want %q", path, wantPath)
	}
	Game.Info("active file output", "mode", "file")
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	stderr := restore()
	data, err := os.ReadFile(wantPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", wantPath, err)
	}
	content := string(data)
	for _, value := range []string{`"msg":"active file output"`, `"category":"game"`, `"mode":"file"`} {
		if !strings.Contains(content, value) {
			t.Fatalf("expected file output to include %s, got %q", value, content)
		}
	}
	if !strings.Contains(stderr, "active file output") {
		t.Fatalf("expected console output while file output was enabled, got %q", stderr)
	}
}

func TestCloseFileOutputReturnsConsoleOnlyLogging(t *testing.T) {
	resetLogging(t)
	restore := captureStderr(t)
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}
	Configure("info")
	directory := t.TempDir()
	path, err := ConfigureFileOutput(directory, "game-server")
	if err != nil {
		t.Fatalf("ConfigureFileOutput() error = %v", err)
	}
	Game.Info("file phase")
	if err := CloseFileOutput(); err != nil {
		t.Fatalf("CloseFileOutput() error = %v", err)
	}
	Game.Info("console phase")
	stderr := restore()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("os.ReadFile(%q) error = %v", path, err)
	}
	content := string(data)
	if !strings.Contains(content, "file phase") || strings.Contains(content, "console phase") {
		t.Fatalf("unexpected file output %q", content)
	}
	if !strings.Contains(stderr, "console phase") {
		t.Fatalf("expected console output after CloseFileOutput, got %q", stderr)
	}
}

func TestConfiguredIdentityAndStatusReachFileOutput(t *testing.T) {
	resetLogging(t)
	restore := captureStderr(t)
	defer restore()
	if err := CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}
	Configure("info")
	path, err := ConfigureFileOutput(t.TempDir(), "game-server")
	if err != nil {
		t.Fatal(err)
	}
	Server.Info("identity record")
	if Status().Degraded {
		t.Fatalf("status=%#v", Status())
	}
	if err := CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, value := range []string{`"service":"game-server"`, `"build_version":"test-build"`, `"environment":"test"`, `"service_instance_id":"550e8400-e29b-41d4-a716-446655440000"`} {
		if !strings.Contains(content, value) {
			t.Fatalf("missing %s in %s", value, content)
		}
	}
	if !Status().Closed {
		t.Fatalf("status=%#v", Status())
	}
}

func TestFileFailureIsReportedAsDegraded(t *testing.T) {
	resetLogging(t)
	restore := captureStderr(t)
	defer restore()
	if err := CloseFileOutput(); err != nil {
		t.Fatal(err)
	}
	if err := ConfigureRuntime(testIdentity()); err != nil {
		t.Fatal(err)
	}
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ConfigureFileOutput(filepath.Join(blocked, "logs"), "game-server"); err != nil {
		t.Fatal(err)
	}
	if !Status().Degraded || Status().FailureCount == 0 {
		t.Fatalf("status=%#v", Status())
	}
}

func TestConfigureRuntimeRejectsIncompleteIdentity(t *testing.T) {
	if err := ConfigureRuntime(servicelog.ServiceIdentity{Name: ServiceName}); err == nil {
		t.Fatal("expected incomplete identity error")
	}
}

func resetLogging(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		if err := CloseFileOutput(); err != nil {
			t.Errorf("cleanup CloseFileOutput() error = %v", err)
		}
		Configure("warn")
	})
}

func captureStderr(t *testing.T) func() string {
	t.Helper()
	original := os.Stderr
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stderr = writer
	closed := false
	t.Cleanup(func() {
		if closed {
			return
		}
		os.Stderr = original
		_ = writer.Close()
		_ = reader.Close()
	})
	return func() string {
		t.Helper()
		if !closed {
			closed = true
			os.Stderr = original
			_ = writer.Close()
		}
		data, _ := io.ReadAll(reader)
		_ = reader.Close()
		return string(data)
	}
}
