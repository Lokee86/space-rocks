package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/config"
)

func logConfig() config.Config {
	return config.Config{ServiceInstanceID: "instance-1", Environment: "test", BuildVersion: "v1", LogLevel: "info"}
}

func TestNewLoggerSinkTogglesAndBaseContext(t *testing.T) {
	cfg := logConfig()
	cfg.ConsoleLogging = true
	cfg.FileLogging = true
	var console, file bytes.Buffer
	logger, err := New(cfg, &console, &file)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("started", "event", "test")
	for name, output := range map[string]string{"console": console.String(), "file": file.String()} {
		var record map[string]any
		if err := json.Unmarshal([]byte(output), &record); err != nil {
			t.Fatalf("%s JSON: %v", name, err)
		}
		for _, field := range []string{"service", "service_instance_id", "environment", "build_version"} {
			if record[field] == nil {
				t.Fatalf("%s missing %s: %s", name, field, output)
			}
		}
	}
}

func TestNewLoggerIndependentSinkToggles(t *testing.T) {
	cfg := logConfig()
	cfg.ConsoleLogging = true
	var console bytes.Buffer
	logger, err := New(cfg, &console, nil)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("console")
	if console.Len() == 0 {
		t.Fatal("console sink did not receive record")
	}
	cfg.ConsoleLogging = false
	cfg.FileLogging = false
	if logger, err = New(cfg, nil, nil); err != nil {
		t.Fatal(err)
	} else {
		logger.Info("discarded")
	}
}

func TestNewLoggerRequiresEnabledSinkWriter(t *testing.T) {
	cfg := logConfig()
	cfg.ConsoleLogging = true
	if _, err := New(cfg, nil, nil); err == nil || !strings.Contains(err.Error(), "console sink") {
		t.Fatalf("error = %v", err)
	}
	cfg.ConsoleLogging = false
	cfg.FileLogging = true
	if _, err := New(cfg, nil, nil); err == nil || !strings.Contains(err.Error(), "file sink") {
		t.Fatalf("error = %v", err)
	}
}

func TestNewLoggerMinimumLevelFiltering(t *testing.T) {
	cfg := logConfig()
	cfg.ConsoleLogging = true
	cfg.LogLevel = "warn"
	var output bytes.Buffer
	logger, err := New(cfg, &output, nil)
	if err != nil {
		t.Fatal(err)
	}
	logger.Info("hidden")
	logger.Warn("visible")
	if strings.Contains(output.String(), "hidden") || !strings.Contains(output.String(), "visible") {
		t.Fatalf("unexpected output: %s", output.String())
	}
	cfg.LogLevel = "trace"
	if _, err := New(cfg, &output, nil); err == nil || !strings.Contains(err.Error(), "invalid minimum log level") {
		t.Fatalf("error = %v", err)
	}
}
