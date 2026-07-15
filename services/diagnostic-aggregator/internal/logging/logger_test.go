package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/observability"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

func testConfig(directory string) Config {
	return Config{
		Identity: servicelog.ServiceIdentity{
			Name:        ServiceName,
			Version:     "test-build",
			Environment: "test",
			InstanceID:  "550e8400-e29b-41d4-a716-446655440002",
		},
		Directory: directory,
		Prefix:    DefaultPrefix,
	}
}

func TestOperationalLoggerIdentityPathAndPolicy(t *testing.T) {
	cfg := testConfig(t.TempDir())
	logger, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("aggregator started")
	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(cfg.ActivePath())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, value := range []string{`"service":"diagnostic-aggregator"`, `"build_version":"test-build"`, `"environment":"test"`, `"service_instance_id":"550e8400-e29b-41d4-a716-446655440002"`} {
		if !strings.Contains(content, value) {
			t.Fatalf("missing %s in %s", value, content)
		}
	}
	runtimeCfg := cfg.runtimeConfig()
	if runtimeCfg.File.SegmentMaxAge != time.Second*time.Duration(observability.FileLoggingMaxActiveSegmentAgeSeconds) || runtimeCfg.File.RetentionMaxAge != time.Second*time.Duration(observability.RetentionDefaultAgeOperationalSeconds) || runtimeCfg.File.CompressionEnabled != observability.FileLoggingCompressionEnabled {
		t.Fatalf("policy=%#v", runtimeCfg.File)
	}
}

func TestOperationalLoggerRecoversAndCompressesInterruptedSegment(t *testing.T) {
	directory := t.TempDir()
	activePath := filepath.Join(directory, DefaultPrefix+".jsonl.open")
	if err := os.WriteFile(activePath, []byte("{\"message\":\"interrupted\"}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	logger, err := Open(testConfig(directory))
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()
	archives, err := filepath.Glob(filepath.Join(directory, "archive", "*", "*.jsonl.gz"))
	if err != nil || len(archives) != 1 {
		t.Fatalf("archives=%v err=%v", archives, err)
	}
}

func TestOperationalLoggerFileFailureIsDegraded(t *testing.T) {
	blocked := filepath.Join(t.TempDir(), "blocked")
	if err := os.WriteFile(blocked, []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	logger, err := Open(testConfig(filepath.Join(blocked, "logs")))
	if err != nil {
		t.Fatal(err)
	}
	defer logger.Close()
	logger.Info("console-only operation")
	if !logger.Status().Degraded || logger.Status().FailureCount == 0 {
		t.Fatalf("status=%#v", logger.Status())
	}
}

func TestOperationalLoggerCloseIsIdempotent(t *testing.T) {
	logger, err := Open(testConfig(t.TempDir()))
	if err != nil {
		t.Fatal(err)
	}
	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}
	if err := logger.Close(); err != nil {
		t.Fatal(err)
	}
	if !logger.Status().Closed {
		t.Fatalf("status=%#v", logger.Status())
	}
}
