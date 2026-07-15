package hosted

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/diagnostics"
	operational "github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/logging"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/storage/jsonlstore"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
	"github.com/google/uuid"
)

const (
	defaultStorageRoot              = "data/diagnostic-reports"
	defaultOperationalLogRoot       = "logs/diagnostic-aggregator"
	defaultMaxRequestBytes    int64 = 4 * 1024 * 1024
)

func conservativeLimits() diagnostics.SubmissionLimits {
	return diagnostics.SubmissionLimits{MaxEmbeddedEvents: 100, MaxUserDescriptionBytes: 4096, MaxFailureMessageBytes: 4096, MaxContextStringBytes: 256}
}

type Config struct {
	Enabled                   bool
	BearerToken               string
	StorageRoot               string
	DiagnosticReportRetention time.Duration
	MaxRequestBytes           int64
	SubmissionLimits          diagnostics.SubmissionLimits
	OperationalLogRoot        string
	BuildVersion              string
	Environment               string
	ServiceInstanceID         string
}

func DefaultConfig() Config {
	return Config{
		StorageRoot:               defaultStorageRoot,
		DiagnosticReportRetention: jsonlstore.DefaultDiagnosticReportRetention,
		MaxRequestBytes:           defaultMaxRequestBytes,
		SubmissionLimits:          conservativeLimits(),
		OperationalLogRoot:        defaultOperationalLogRoot,
		BuildVersion:              "development",
		Environment:               "development",
		ServiceInstanceID:         uuid.NewString(),
	}
}

func LoadConfig() (Config, error) {
	cfg := DefaultConfig()
	var err error
	if value := os.Getenv("DIAGNOSTIC_AGGREGATOR_ENABLED"); value != "" {
		cfg.Enabled, err = strconv.ParseBool(value)
		if err != nil {
			return Config{}, fmt.Errorf("hosted: invalid enabled")
		}
	}
	if value := os.Getenv("DIAGNOSTIC_AGGREGATOR_TOKEN"); value != "" {
		cfg.BearerToken = value
	}
	if value := os.Getenv("DIAGNOSTIC_AGGREGATOR_STORAGE_ROOT"); value != "" {
		cfg.StorageRoot = value
	}
	if value := os.Getenv("DIAGNOSTIC_AGGREGATOR_LOG_ROOT"); value != "" {
		cfg.OperationalLogRoot = value
	}
	if value := os.Getenv("BUILD_VERSION"); value != "" {
		cfg.BuildVersion = value
	}
	if value := os.Getenv("ENVIRONMENT"); value != "" {
		cfg.Environment = value
	}
	if value := os.Getenv("DIAGNOSTIC_AGGREGATOR_RETENTION"); value != "" {
		cfg.DiagnosticReportRetention, err = time.ParseDuration(value)
		if err != nil {
			return Config{}, fmt.Errorf("hosted: invalid report retention")
		}
	}
	if value := os.Getenv("DIAGNOSTIC_AGGREGATOR_MAX_REQUEST_BYTES"); value != "" {
		cfg.MaxRequestBytes, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("hosted: invalid max request bytes")
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.StorageRoot) == "" {
		return errors.New("hosted: storage root is required")
	}
	if c.DiagnosticReportRetention <= 0 || c.MaxRequestBytes <= 0 {
		return errors.New("hosted: numeric and duration limits must be positive")
	}
	if err := c.SubmissionLimits.Validate(); err != nil {
		return err
	}
	if c.Enabled && (strings.TrimSpace(c.BearerToken) == "" || strings.IndexAny(c.BearerToken, " \t\r\n") >= 0) {
		return errors.New("hosted: bearer token must be nonblank and whitespace-free")
	}
	if c.Enabled {
		if err := operationalLogConfig(c).Validate(); err != nil {
			return fmt.Errorf("hosted: operational logging: %w", err)
		}
	}
	return nil
}

func reportStoreConfig(c Config) jsonlstore.Config {
	cfg := jsonlstore.DefaultConfig(c.StorageRoot)
	cfg.DiagnosticReportRetention = c.DiagnosticReportRetention
	return cfg
}

func operationalLogConfig(c Config) operational.Config {
	return operational.Config{
		Identity: servicelog.ServiceIdentity{
			Name:        operational.ServiceName,
			Version:     strings.TrimSpace(c.BuildVersion),
			Environment: strings.TrimSpace(c.Environment),
			InstanceID:  strings.TrimSpace(c.ServiceInstanceID),
		},
		Directory: c.OperationalLogRoot,
		Prefix:    operational.DefaultPrefix,
	}
}
