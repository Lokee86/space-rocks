package config

import (
	"strings"
	"testing"
	"time"
)

func mapEnv(values map[string]string) EnvLookup {
	return func(key string) (string, bool) { value, ok := values[key]; return value, ok }
}

func fixedUUID() (string, error) { return "123e4567-e89b-42d3-a456-426614174000", nil }

func TestLoadWithDefaultsAndInjectedUUID(t *testing.T) {
	cfg, err := LoadWith(mapEnv(nil), fixedUUID)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddress != defaultListenAddress || cfg.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if !cfg.ConsoleLogging || !cfg.FileLogging || cfg.LogDirectory != defaultLogDirectory || cfg.LogLevel != "info" {
		t.Fatalf("unexpected logging defaults: %+v", cfg)
	}
	if cfg.ServiceInstanceID != "123e4567-e89b-42d3-a456-426614174000" {
		t.Fatalf("unexpected UUID: %s", cfg.ServiceInstanceID)
	}
}

func TestLoadWithOverridesReadHeaderTimeout(t *testing.T) {
	cfg, err := LoadWith(mapEnv(map[string]string{
		"LOG_AGGREGATOR_LISTEN_ADDRESS":      "0.0.0.0:9090",
		"LOG_AGGREGATOR_READ_HEADER_TIMEOUT": "2s",
		"LOG_AGGREGATOR_READ_TIMEOUT":        "3s",
		"LOG_AGGREGATOR_FILE_LOGGING":        "false",
		"LOG_AGGREGATOR_LOG_LEVEL":           "WARN",
	}), fixedUUID)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ListenAddress != "0.0.0.0:9090" || cfg.ReadHeaderTimeout != 2*time.Second || cfg.ReadTimeout != 3*time.Second || cfg.FileLogging || cfg.LogLevel != "warn" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadWithRejectsInvalidValues(t *testing.T) {
	cases := []struct{ key, value, want string }{
		{"LOG_AGGREGATOR_LISTEN_ADDRESS", "localhost", "listen address"},
		{"LOG_AGGREGATOR_READ_HEADER_TIMEOUT", "nope", "positive duration"},
		{"LOG_AGGREGATOR_LOG_LEVEL", "trace", "invalid log level"},
		{"LOG_AGGREGATOR_SERVICE_INSTANCE_ID", "not-a-uuid", "valid UUID"},
	}
	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			_, err := LoadWith(mapEnv(map[string]string{tc.key: tc.value}), fixedUUID)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
		})
	}
}

func TestLoadWithRequiresInjectedDependencies(t *testing.T) {
	if _, err := LoadWith(nil, fixedUUID); err == nil || !strings.Contains(err.Error(), "environment lookup") {
		t.Fatal(err)
	}
	if _, err := LoadWith(mapEnv(nil), nil); err == nil || !strings.Contains(err.Error(), "UUID generator") {
		t.Fatal(err)
	}
}
