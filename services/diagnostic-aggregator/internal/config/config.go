package config

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/serviceidentity"
)

const (
	defaultListenAddress             = "127.0.0.1:8091"
	defaultEnvironment               = "development"
	defaultBuildVersion              = "dev"
	defaultLogDirectory              = serviceidentity.DefaultLogDirectory
	defaultDiagnosticReportRetention = 14 * 24 * time.Hour
)

type Config struct {
	ListenAddress             string
	Environment               string
	BuildVersion              string
	ServiceInstanceID         string
	ReadHeaderTimeout         time.Duration
	ReadTimeout               time.Duration
	WriteTimeout              time.Duration
	IdleTimeout               time.Duration
	ShutdownTimeout           time.Duration
	ConsoleLogging            bool
	FileLogging               bool
	LogLevel                  string
	LogDirectory              string
	DiagnosticReportRetention time.Duration
}

type EnvLookup func(string) (string, bool)
type UUIDGenerator func() (string, error)

func Load() (Config, error) { return LoadWith(os.LookupEnv, newUUID) }

func LoadWith(getenv EnvLookup, generateUUID UUIDGenerator) (Config, error) {
	if getenv == nil || generateUUID == nil {
		return Config{}, fmt.Errorf("config: environment lookup and UUID generator are required")
	}
	c := Config{
		ListenAddress:             env(getenv, "LISTEN_ADDRESS", defaultListenAddress),
		Environment:               env(getenv, "ENVIRONMENT", defaultEnvironment),
		BuildVersion:              env(getenv, "BUILD_VERSION", defaultBuildVersion),
		ReadHeaderTimeout:         5 * time.Second,
		ReadTimeout:               15 * time.Second,
		WriteTimeout:              15 * time.Second,
		IdleTimeout:               60 * time.Second,
		ShutdownTimeout:           10 * time.Second,
		ConsoleLogging:            true,
		FileLogging:               true,
		LogLevel:                  "info",
		LogDirectory:              defaultLogDirectory,
		DiagnosticReportRetention: defaultDiagnosticReportRetention,
	}
	var err error
	for _, setting := range []struct {
		name   string
		target *time.Duration
	}{
		{"READ_HEADER_TIMEOUT", &c.ReadHeaderTimeout}, {"READ_TIMEOUT", &c.ReadTimeout},
		{"WRITE_TIMEOUT", &c.WriteTimeout}, {"IDLE_TIMEOUT", &c.IdleTimeout}, {"SHUTDOWN_TIMEOUT", &c.ShutdownTimeout},
		{"DIAGNOSTIC_REPORT_RETENTION", &c.DiagnosticReportRetention},
	} {
		if *setting.target, err = duration(getenv, setting.name, *setting.target); err != nil {
			return Config{}, err
		}
	}
	if c.ConsoleLogging, err = boolean(getenv, "CONSOLE_LOGGING", c.ConsoleLogging); err != nil {
		return Config{}, err
	}
	if c.FileLogging, err = boolean(getenv, "FILE_LOGGING", c.FileLogging); err != nil {
		return Config{}, err
	}
	c.LogLevel = strings.ToLower(env(getenv, "LOG_LEVEL", c.LogLevel))
	c.LogDirectory = env(getenv, "LOG_DIRECTORY", c.LogDirectory)
	c.ServiceInstanceID = env(getenv, "SERVICE_INSTANCE_ID", "")
	if c.ServiceInstanceID == "" {
		c.ServiceInstanceID, err = generateUUID()
		if err != nil {
			return Config{}, fmt.Errorf("config: generate service instance UUID: %w", err)
		}
	}
	if err := validate(c); err != nil {
		return Config{}, err
	}
	return c, nil
}

func env(getenv EnvLookup, name, fallback string) string {
	if value, ok := getenv(serviceidentity.EnvPrefix + name); ok {
		return strings.TrimSpace(value)
	}
	if value, ok := getenv(serviceidentity.LegacyEnvPrefix + name); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func duration(getenv EnvLookup, name string, fallback time.Duration) (time.Duration, error) {
	value := env(getenv, name, "")
	if value == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("config: %s%s must be a positive duration: %q", serviceidentity.EnvPrefix, name, value)
	}
	return d, nil
}

func boolean(getenv EnvLookup, name string, fallback bool) (bool, error) {
	value := env(getenv, name, "")
	if value == "" {
		return fallback, nil
	}
	result, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("config: %s%s must be boolean: %q", serviceidentity.EnvPrefix, name, value)
	}
	return result, nil
}

func validate(c Config) error {
	if strings.TrimSpace(c.Environment) == "" {
		return fmt.Errorf("config: environment is required")
	}
	if strings.TrimSpace(c.BuildVersion) == "" {
		return fmt.Errorf("config: build version is required")
	}
	if _, err := net.ResolveTCPAddr("tcp", c.ListenAddress); err != nil {
		return fmt.Errorf("config: invalid listen address %q: %w", c.ListenAddress, err)
	}
	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return fmt.Errorf("config: invalid log level %q", c.LogLevel)
	}
	if strings.TrimSpace(c.LogDirectory) == "" {
		return fmt.Errorf("config: log directory is required")
	}
	if !validUUID(c.ServiceInstanceID) {
		return fmt.Errorf("config: service instance ID must be a valid UUID: %q", c.ServiceInstanceID)
	}
	return nil
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return false
	}
	for i, r := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			continue
		}
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return value[14] == '4' && strings.ContainsRune("89abAB", rune(value[19]))
}

func newUUID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", bytes[:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:]), nil
}
