package hosted

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Enabled || cfg.DiagnosticReportRetention != 14*24*time.Hour || cfg.MaxRequestBytes != 4*1024*1024 || cfg.OperationalLogRoot == "" || cfg.BuildVersion == "" || cfg.Environment == "" || cfg.ServiceInstanceID == "" {
		t.Fatalf("defaults = %#v", cfg)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestLoadConfigDisabledDoesNotRequireToken(t *testing.T) {
	for _, name := range []string{"DIAGNOSTIC_AGGREGATOR_ENABLED", "DIAGNOSTIC_AGGREGATOR_TOKEN"} {
		t.Setenv(name, "")
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Enabled {
		t.Fatal("default must be disabled")
	}
}

func TestLoadConfigRetentionOverride(t *testing.T) {
	t.Setenv("DIAGNOSTIC_AGGREGATOR_RETENTION", "336h")
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DiagnosticReportRetention != 14*24*time.Hour {
		t.Fatalf("retention = %s, want %s", cfg.DiagnosticReportRetention, 14*24*time.Hour)
	}
}

func TestLoadConfigRejectsInvalidRetention(t *testing.T) {
	t.Setenv("DIAGNOSTIC_AGGREGATOR_RETENTION", "not-a-duration")
	if _, err := LoadConfig(); err == nil {
		t.Fatal("expected invalid retention duration error")
	}
}

func TestEnabledConfigRequiresSafeToken(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	for _, token := range []string{"", "  secret", "secret token"} {
		cfg.BearerToken = token
		if err := cfg.Validate(); err == nil {
			t.Fatalf("token %q accepted", token)
		}
	}
}
