package jsonlstore

import (
	"errors"
	"fmt"
	"time"
)

// DefaultDiagnosticReportRetention is the local default until the shared
// retention contract exposes a generated duration for this service.
const DefaultDiagnosticReportRetention = 14 * 24 * time.Hour

// Config controls diagnostic-report JSONL storage.
type Config struct {
	Root                      string
	SegmentMaxBytes           int64
	SegmentMaxAge             time.Duration
	DiagnosticReportRetention time.Duration
	Compression               bool
	FlushInterval             time.Duration
}

// DefaultConfig returns the agreed production defaults rooted at root.
func DefaultConfig(root string) Config {
	return Config{
		Root:                      root,
		SegmentMaxBytes:           64 * 1024 * 1024,
		SegmentMaxAge:             time.Hour,
		DiagnosticReportRetention: DefaultDiagnosticReportRetention,
		Compression:               true,
		FlushInterval:             time.Second,
	}
}

// NewConfig returns a validated configuration using the backend defaults.
func NewConfig(root string) (Config, error) {
	config := DefaultConfig(root)
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

// Validate checks the configuration values owned by report storage.
func (config Config) Validate() error {
	if config.Root == "" {
		return errors.New("jsonlstore: storage root is required")
	}
	if config.SegmentMaxBytes <= 0 {
		return fmt.Errorf("jsonlstore: segment max bytes must be greater than zero, got %d", config.SegmentMaxBytes)
	}
	if config.SegmentMaxAge <= 0 {
		return fmt.Errorf("jsonlstore: segment max age must be greater than zero, got %s", config.SegmentMaxAge)
	}
	if config.DiagnosticReportRetention <= 0 {
		return fmt.Errorf("jsonlstore: diagnostic report retention must be greater than zero, got %s", config.DiagnosticReportRetention)
	}
	if config.FlushInterval < 0 {
		return fmt.Errorf("jsonlstore: flush interval cannot be negative, got %s", config.FlushInterval)
	}
	return nil
}
