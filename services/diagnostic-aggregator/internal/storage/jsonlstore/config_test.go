package jsonlstore

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("storage")
	if config.Root != "storage" || config.SegmentMaxBytes != 64*1024*1024 || config.SegmentMaxAge != time.Hour {
		t.Fatalf("segment defaults = %#v", config)
	}
	if config.RetentionMaxAge != 14*24*time.Hour || config.RetentionMaxBytes != 1*1024*1024*1024 {
		t.Fatalf("retention defaults = %#v", config)
	}
	if !config.Compression || config.FlushInterval != time.Second {
		t.Fatalf("compression/flush defaults = %#v", config)
	}
	if err := config.Validate(); err != nil {
		t.Fatalf("default config is invalid: %v", err)
	}
}

func TestConfigValidation(t *testing.T) {
	base := DefaultConfig("storage")
	tests := []struct {
		name string
		edit func(*Config)
		want string
	}{
		{"root", func(config *Config) { config.Root = "" }, "storage root"},
		{"segment bytes", func(config *Config) { config.SegmentMaxBytes = 0 }, "segment max bytes"},
		{"segment age", func(config *Config) { config.SegmentMaxAge = 0 }, "segment max age"},
		{"retention age", func(config *Config) { config.RetentionMaxAge = 0 }, "retention max age"},
		{"retention bytes", func(config *Config) { config.RetentionMaxBytes = 0 }, "retention max bytes"},
		{"negative flush", func(config *Config) { config.FlushInterval = -time.Second }, "flush interval"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := base
			test.edit(&config)
			err := config.Validate()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want text %q", err, test.want)
			}
		})
	}
}

func TestZeroFlushIntervalIsValid(t *testing.T) {
	config := DefaultConfig("storage")
	config.FlushInterval = 0
	if err := config.Validate(); err != nil {
		t.Fatalf("zero flush interval should disable periodic flushing: %v", err)
	}
}
