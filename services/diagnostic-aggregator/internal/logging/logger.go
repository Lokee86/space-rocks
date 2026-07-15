package logging

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

const (
	ServiceName   = observability.ServiceNameDiagnosticAggregator
	DefaultPrefix = observability.ServiceNameDiagnosticAggregator

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
	emitter *observability.Emitter
}

func Open(config Config) (*Logger, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	runtime, err := servicelog.Open(config.runtimeConfig())
	if err != nil {
		return nil, err
	}
	emitter, err := observability.New(observability.Config{
		Service: observability.ServiceKeyDiagnosticAggregator, Environment: config.Identity.Environment,
		BuildVersion: config.Identity.Version, ServiceInstanceID: config.Identity.InstanceID,
		PID: os.Getpid(), Sink: runtime, WarningWriter: os.Stderr,
	})
	if err != nil {
		_ = runtime.Close()
		return nil, err
	}
	return &Logger{runtime: runtime, emitter: emitter}, nil
}

func (l *Logger) Emit(request observability.Request) observability.Result {
	if l == nil || l.emitter == nil {
		return observability.Result{}
	}
	return l.emitter.Emit(request)
}

func (l *Logger) Status() servicelog.Status {
	if l == nil || l.runtime == nil {
		return servicelog.Status{}
	}
	return l.runtime.Status()
}

func (l *Logger) EventStatus() observability.Status {
	if l == nil || l.emitter == nil {
		return observability.Status{}
	}
	return l.emitter.Status()
}

func (l *Logger) Close() error {
	if l == nil || l.runtime == nil {
		return nil
	}
	return l.runtime.Close()
}
