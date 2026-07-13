package aggregatorclient

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultEnabled        = false
	defaultEndpointURL    = ""
	defaultBearerToken    = ""
	defaultQueueCapacity  = 1024
	defaultBatchSize      = 64
	defaultFlushInterval  = 1 * time.Second
	defaultRequestTimeout = 5 * time.Second
	defaultSpoolDirectory = "var/spool/space-rocks/observability"
	defaultSpoolByteCap   = 64 * 1024 * 1024
	defaultSpoolEnabled   = true
)

// Config contains the independently controlled aggregator sink configuration.
type Config struct {
	Enabled        bool
	EndpointURL    string
	BearerToken    string
	QueueCapacity  int
	BatchSize      int
	FlushInterval  time.Duration
	RequestTimeout time.Duration
	SpoolDirectory string
	SpoolByteCap   int64
	SpoolEnabled   bool
}

// DefaultConfig returns stable defaults for tests and technical-release deployments.
func DefaultConfig() Config {
	return Config{
		Enabled:        defaultEnabled,
		EndpointURL:    defaultEndpointURL,
		BearerToken:    defaultBearerToken,
		QueueCapacity:  defaultQueueCapacity,
		BatchSize:      defaultBatchSize,
		FlushInterval:  defaultFlushInterval,
		RequestTimeout: defaultRequestTimeout,
		SpoolDirectory: defaultSpoolDirectory,
		SpoolByteCap:   defaultSpoolByteCap,
		SpoolEnabled:   defaultSpoolEnabled,
	}
}

// ConfigFromEnv loads OBS_AGGREGATOR_* overrides and validates the result.
func ConfigFromEnv() (Config, error) {
	config := DefaultConfig()
	var err error
	if config.Enabled, err = envBool("OBS_AGGREGATOR_ENABLED", config.Enabled); err != nil {
		return Config{}, err
	}
	if config.EndpointURL, err = envString("OBS_AGGREGATOR_ENDPOINT_URL", config.EndpointURL); err != nil {
		return Config{}, err
	}
	if config.BearerToken, err = envString("OBS_AGGREGATOR_BEARER_TOKEN", config.BearerToken); err != nil {
		return Config{}, err
	}
	if config.QueueCapacity, err = envInt("OBS_AGGREGATOR_QUEUE_CAPACITY", config.QueueCapacity); err != nil {
		return Config{}, err
	}
	if config.BatchSize, err = envInt("OBS_AGGREGATOR_BATCH_SIZE", config.BatchSize); err != nil {
		return Config{}, err
	}
	if config.FlushInterval, err = envDuration("OBS_AGGREGATOR_FLUSH_INTERVAL", config.FlushInterval); err != nil {
		return Config{}, err
	}
	if config.RequestTimeout, err = envDuration("OBS_AGGREGATOR_REQUEST_TIMEOUT", config.RequestTimeout); err != nil {
		return Config{}, err
	}
	if config.SpoolDirectory, err = envString("OBS_AGGREGATOR_SPOOL_DIRECTORY", config.SpoolDirectory); err != nil {
		return Config{}, err
	}
	if config.SpoolByteCap, err = envInt64("OBS_AGGREGATOR_SPOOL_BYTE_CAP", config.SpoolByteCap); err != nil {
		return Config{}, err
	}
	if config.SpoolEnabled, err = envBool("OBS_AGGREGATOR_SPOOL_ENABLED", config.SpoolEnabled); err != nil {
		return Config{}, err
	}
	return config, config.Validate()
}

func (c Config) Validate() error {
	if c.QueueCapacity <= 0 {
		return fmt.Errorf("queue capacity must be positive")
	}
	if c.BatchSize <= 0 {
		return fmt.Errorf("batch size must be positive")
	}
	if c.FlushInterval <= 0 {
		return fmt.Errorf("flush interval must be positive")
	}
	if c.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}
	if c.SpoolByteCap <= 0 {
		return fmt.Errorf("spool byte cap must be positive")
	}
	if !c.Enabled {
		return nil
	}
	if strings.TrimSpace(c.EndpointURL) == "" {
		return fmt.Errorf("endpoint URL is required when aggregator is enabled")
	}
	parsed, err := url.ParseRequestURI(c.EndpointURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("endpoint URL must be an absolute HTTP(S) URL")
	}
	return nil
}

func envString(name, fallback string) (string, error) {
	if value, ok := os.LookupEnv(name); ok {
		return value, nil
	}
	return fallback, nil
}
func envBool(name string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s: %w", name, err)
	}
	return parsed, nil
}
func envInt(name string, fallback int) (int, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return parsed, nil
}
func envInt64(name string, fallback int64) (int64, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return parsed, nil
}
func envDuration(name string, fallback time.Duration) (time.Duration, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	return parsed, nil
}
