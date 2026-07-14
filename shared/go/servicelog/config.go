// Package servicelog defines the shared configuration boundary for service-owned
// structured logging. It does not provide a writer or impose project policy.
package servicelog

import (
	"errors"
	"strings"
	"time"
)

// Config contains caller-supplied service logging identity and output policy.
type Config struct {
	Identity ServiceIdentity
	File     FilePolicy
	Flush    FlushPolicy

	ConsoleEnabled bool
	FileEnabled    bool
}

// ServiceIdentity identifies the service emitting a log record.
type ServiceIdentity struct {
	Name        string
	Version     string
	Environment string
	InstanceID  string
}

// FilePolicy describes the caller-selected location and naming policy for file
// output, including caller-supplied rolling and retention policy.
type FilePolicy struct {
	Directory          string
	Prefix             string
	SegmentMaxBytes    int64
	SegmentMaxAge      time.Duration
	RetentionMaxAge    time.Duration
	RetentionMaxBytes  int64
	CompressionEnabled bool
}

// FlushPolicy describes when buffered file output should be flushed.
type FlushPolicy struct {
	Interval time.Duration
}

// Validate reports configuration values that cannot be used safely.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Identity.Name) == "" {
		return errors.New("servicelog: identity name is required")
	}
	if c.Flush.Interval < 0 {
		return errors.New("servicelog: flush interval cannot be negative")
	}
	if c.FileEnabled {
		if strings.TrimSpace(c.File.Directory) == "" {
			return errors.New("servicelog: file directory is required when file output is enabled")
		}
		if strings.TrimSpace(c.File.Prefix) == "" {
			return errors.New("servicelog: file prefix is required when file output is enabled")
		}
		if c.File.SegmentMaxBytes <= 0 {
			return errors.New("servicelog: segment max bytes must be positive when file output is enabled")
		}
		if c.File.SegmentMaxAge <= 0 {
			return errors.New("servicelog: segment max age must be positive when file output is enabled")
		}
		if c.File.RetentionMaxAge <= 0 {
			return errors.New("servicelog: retention max age must be positive when file output is enabled")
		}
		if c.File.RetentionMaxBytes <= 0 {
			return errors.New("servicelog: retention max bytes must be positive when file output is enabled")
		}
	}
	return nil
}
