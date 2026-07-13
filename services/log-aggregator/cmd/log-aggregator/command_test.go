package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/config"
)

type testApp struct{ err error }

func (a testApp) Run(context.Context) error { return a.err }

type testWriter struct {
	bytes.Buffer
	closed   int
	closeErr error
}

func (w *testWriter) Close() error { w.closed++; return w.closeErr }

func commandConfig() config.Config {
	return config.Config{ServiceInstanceID: "i", Environment: "test", BuildVersion: "v1", LogLevel: "info", ConsoleLogging: true}
}
func noFile(_ config.Config) (io.WriteCloser, error)       { return nil, errors.New("file unavailable") }
func appOK(config.Config, *slog.Logger) (AppRunner, error) { return testApp{}, nil }

func TestRunConfigurationFailure(t *testing.T) {
	var stderr bytes.Buffer
	code := run(context.Background(), &stderr, func() (config.Config, error) { return config.Config{}, errors.New("secret") }, nil, appOK)
	if code != exitConfig {
		t.Fatalf("exit = %d", code)
	}
	var record map[string]string
	if err := json.Unmarshal(stderr.Bytes(), &record); err != nil {
		t.Fatal(err)
	}
	if record["event"] != "service_startup_failed" || record["service"] != "log-aggregator" || record["error_code"] != "configuration_invalid" || strings.Contains(stderr.String(), "secret") {
		t.Fatalf("unexpected bootstrap record: %s", stderr.String())
	}
}

func TestRunDisabledSinks(t *testing.T) {
	cfg := commandConfig()
	cfg.ConsoleLogging = false
	called := false
	code := run(context.Background(), io.Discard, func() (config.Config, error) { return cfg, nil }, func(config.Config) (io.WriteCloser, error) { called = true; return nil, nil }, appOK)
	if code != exitSuccess || called {
		t.Fatalf("exit=%d opener_called=%v", code, called)
	}
}

func TestRunDegradesWhenServiceLogUnavailable(t *testing.T) {
	cfg := commandConfig()
	cfg.FileLogging = true
	var stderr bytes.Buffer
	code := run(context.Background(), &stderr, func() (config.Config, error) { return cfg, nil }, noFile, appOK)
	if code != exitSuccess || !strings.Contains(stderr.String(), "environment_degraded") || !strings.Contains(stderr.String(), "service_log_file_unavailable") {
		t.Fatalf("exit=%d output=%s", code, stderr.String())
	}
}

func TestRunAppFactoryFailureClosesServiceLog(t *testing.T) {
	cfg := commandConfig()
	cfg.FileLogging = true
	writer := &testWriter{}
	code := run(context.Background(), io.Discard, func() (config.Config, error) { return cfg, nil }, func(config.Config) (io.WriteCloser, error) { return writer, nil }, func(config.Config, *slog.Logger) (AppRunner, error) { return nil, errors.New("factory") })
	if code != exitFailure || writer.closed != 1 {
		t.Fatalf("exit=%d closed=%d", code, writer.closed)
	}
}

func TestRunRuntimeFailure(t *testing.T) {
	cfg := commandConfig()
	code := run(context.Background(), io.Discard, func() (config.Config, error) { return cfg, nil }, nil, func(config.Config, *slog.Logger) (AppRunner, error) { return testApp{err: errors.New("runtime")}, nil })
	if code != exitFailure {
		t.Fatalf("exit = %d", code)
	}
}

func TestRunSuccessAndCloseFailure(t *testing.T) {
	cfg := commandConfig()
	cfg.FileLogging = true
	writer := &testWriter{closeErr: errors.New("close")}
	var stderr bytes.Buffer
	code := run(context.Background(), &stderr, func() (config.Config, error) { return cfg, nil }, func(config.Config) (io.WriteCloser, error) { return writer, nil }, appOK)
	if code != exitFailure || writer.closed != 1 || !strings.Contains(stderr.String(), "service_log_close_failed") {
		t.Fatalf("exit=%d closed=%d output=%s", code, writer.closed, stderr.String())
	}
	writer = &testWriter{}
	code = run(context.Background(), io.Discard, func() (config.Config, error) { return cfg, nil }, func(config.Config) (io.WriteCloser, error) { return writer, nil }, appOK)
	if code != exitSuccess || writer.closed != 1 {
		t.Fatalf("exit=%d closed=%d", code, writer.closed)
	}
}

func TestRunNilDependenciesAreSafeBootstrapFailures(t *testing.T) {
	if code := run(context.Background(), nil, nil, nil, appOK); code != exitConfig {
		t.Fatalf("nil loader exit = %d", code)
	}
	cfg := commandConfig()
	var stderr bytes.Buffer
	if code := run(context.Background(), &stderr, func() (config.Config, error) { return cfg, nil }, nil, nil); code != exitFailure {
		t.Fatalf("nil factory exit = %d", code)
	}
	if !strings.Contains(stderr.String(), "runtime_build_failed") {
		t.Fatalf("output = %s", stderr.String())
	}
}

func TestRunClosesWriterReturnedWithOpenError(t *testing.T) {
	cfg := commandConfig()
	cfg.FileLogging = true
	writer := &testWriter{}
	code := run(context.Background(), io.Discard, func() (config.Config, error) { return cfg, nil }, func(config.Config) (io.WriteCloser, error) { return writer, errors.New("open") }, appOK)
	if code != exitSuccess || writer.closed != 1 {
		t.Fatalf("exit=%d closed=%d", code, writer.closed)
	}
}
