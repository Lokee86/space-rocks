package logging

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

const (
	EnvGlobalLevel = "LOG_LEVEL"
)

const (
	levelOff slog.Level = slog.LevelError + 1

	ServiceName                 = observability.ServiceNameGameServer
	fileSegmentMaxBytes   int64 = 64 * 1024 * 1024
	fileRetentionMaxBytes int64 = 1 * 1024 * 1024 * 1024
)

var (
	rootLevel    = new(slog.LevelVar)
	logRuntime   *servicelog.Runtime
	eventEmitter *observability.Emitter
	identity     servicelog.ServiceIdentity
	lastStatus   servicelog.Status
)

// Emit is the semantic entry point for game-server-owned canonical events.
func Emit(request observability.Request) observability.Result {
	if eventEmitter == nil {
		eventEmitter = fallbackEmitter()
	}
	return eventEmitter.Emit(request)
}

func init() {
	rootLevel.Set(slog.LevelWarn)
	eventEmitter = fallbackEmitter()
}

// Configure retains the process-level logging configuration entry point. The
// canonical emitter uses generated event levels and categories directly.
func Configure(configuredLevel string) {
	rootLevel.Set(parseLevel(configuredLevel))
}

func ConfigureRuntime(configuredIdentity servicelog.ServiceIdentity) error {
	if err := validateIdentity(configuredIdentity); err != nil {
		return err
	}
	runtime, err := servicelog.Open(servicelog.Config{Identity: configuredIdentity, ConsoleEnabled: true})
	if err != nil {
		return err
	}
	return replaceRuntime(runtime, configuredIdentity)
}

func ConfigureFileOutput(baseDir string, prefix string) (string, error) {
	if err := validateIdentity(identity); err != nil {
		return "", err
	}
	runtime, err := servicelog.Open(fileRuntimeConfig(identity, baseDir, prefix))
	if err != nil {
		return "", err
	}
	if err := replaceRuntime(runtime, identity); err != nil {
		return activeFilePath(baseDir, prefix), err
	}
	return activeFilePath(baseDir, prefix), nil
}

func CloseFileOutput() error {
	oldRuntime := logRuntime
	logRuntime = nil
	eventEmitter = fallbackEmitter()
	if oldRuntime == nil {
		return nil
	}
	err := oldRuntime.Close()
	lastStatus = oldRuntime.Status()
	return err
}

func Status() servicelog.Status {
	if logRuntime == nil {
		return lastStatus
	}
	return logRuntime.Status()
}

func EventStatus() observability.Status {
	if eventEmitter == nil {
		return observability.Status{}
	}
	return eventEmitter.Status()
}

func replaceRuntime(runtime *servicelog.Runtime, configuredIdentity servicelog.ServiceIdentity) error {
	emitter, err := newEmitter(runtime, configuredIdentity)
	if err != nil {
		_ = runtime.Close()
		return err
	}
	oldRuntime := logRuntime
	logRuntime = runtime
	eventEmitter = emitter
	identity = configuredIdentity
	lastStatus = runtime.Status()
	if oldRuntime != nil {
		return oldRuntime.Close()
	}
	return nil
}

func parseLevel(configuredLevel string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(configuredLevel)) {
	case "":
		return slog.LevelWarn
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "off":
		return levelOff
	default:
		return slog.LevelInfo
	}
}

func newEmitter(sink observability.Sink, configuredIdentity servicelog.ServiceIdentity) (*observability.Emitter, error) {
	return observability.New(observability.Config{
		Service: observability.ServiceKeyGameServer, Environment: configuredIdentity.Environment,
		BuildVersion: configuredIdentity.Version, ServiceInstanceID: configuredIdentity.InstanceID,
		PID: os.Getpid(), Sink: sink, WarningWriter: os.Stderr,
	})
}

type stderrSink struct{}

func (stderrSink) WriteRecord(_ []byte, consoleLine string) error {
	_, err := fmt.Fprintln(os.Stderr, consoleLine)
	return err
}

func fallbackEmitter() *observability.Emitter {
	emitter, _ := observability.New(observability.Config{
		Service: observability.ServiceKeyGameServer, Environment: "unconfigured", BuildVersion: "unknown",
		ServiceInstanceID: "00000000-0000-4000-8000-000000000000", PID: os.Getpid(), Sink: stderrSink{}, WarningWriter: os.Stderr,
	})
	return emitter
}

func validateIdentity(value servicelog.ServiceIdentity) error {
	if strings.TrimSpace(value.Name) == "" || strings.TrimSpace(value.Version) == "" || strings.TrimSpace(value.Environment) == "" || strings.TrimSpace(value.InstanceID) == "" {
		return errors.New("game-server logging: service name, build version, environment, and instance ID are required")
	}
	return nil
}

func fileRuntimeConfig(configuredIdentity servicelog.ServiceIdentity, baseDir string, prefix string) servicelog.Config {
	return servicelog.Config{
		Identity: configuredIdentity,
		File: servicelog.FilePolicy{
			Directory:          baseDir,
			Prefix:             prefix,
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

func activeFilePath(baseDir string, prefix string) string {
	return filepath.Join(baseDir, prefix+".jsonl.open")
}
