package logging

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/observability"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

const (
	ServiceName   = "diagnostic-aggregator"
	DefaultPrefix = "diagnostic-aggregator"

	fileSegmentMaxBytes   int64 = 32 * 1024 * 1024
	fileRetentionMaxBytes int64 = 512 * 1024 * 1024
)

type Config struct {
	Identity  servicelog.ServiceIdentity
	Directory string
	Prefix    string
}

func (c Config) Validate() error {
	identity := c.Identity
	if strings.TrimSpace(identity.Name) == "" || strings.TrimSpace(identity.Version) == "" || strings.TrimSpace(identity.Environment) == "" || strings.TrimSpace(identity.InstanceID) == "" {
		return errors.New("diagnostic-aggregator logging: service name, build version, environment, and instance ID are required")
	}
	return c.runtimeConfig().Validate()
}

func (c Config) ActivePath() string {
	return filepath.Join(c.Directory, c.Prefix+".jsonl.open")
}

func (c Config) runtimeConfig() servicelog.Config {
	return servicelog.Config{
		Identity: c.Identity,
		File: servicelog.FilePolicy{
			Directory:          c.Directory,
			Prefix:             c.Prefix,
			SegmentMaxBytes:    fileSegmentMaxBytes,
			SegmentMaxAge:      time.Second * time.Duration(observability.FileLoggingMaxActiveSegmentAgeSeconds),
			RetentionMaxAge:    time.Second * time.Duration(observability.RetentionDefaultAgeOperationalSeconds),
			RetentionMaxBytes:  fileRetentionMaxBytes,
			CompressionEnabled: observability.FileLoggingCompressionEnabled,
		},
		ConsoleEnabled: true,
		FileEnabled:    true,
	}
}

type Logger struct {
	runtime *servicelog.Runtime
}

func Open(config Config) (*Logger, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	runtime, err := servicelog.Open(config.runtimeConfig())
	if err != nil {
		return nil, err
	}
	return &Logger{runtime: runtime}, nil
}

func (l *Logger) Info(message string, args ...any) {
	if l != nil && l.runtime != nil {
		l.runtime.Logger().Log(context.Background(), slog.LevelInfo, message, args...)
	}
}

func (l *Logger) Error(message string, err error, args ...any) {
	if l != nil && l.runtime != nil {
		l.runtime.Logger().Log(context.Background(), slog.LevelError, message, append(args, "error", err)...)
	}
}

func (l *Logger) Status() servicelog.Status {
	if l == nil || l.runtime == nil {
		return servicelog.Status{}
	}
	return l.runtime.Status()
}

func (l *Logger) Close() error {
	if l == nil || l.runtime == nil {
		return nil
	}
	return l.runtime.Close()
}
