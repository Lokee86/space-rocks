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
	CategoryHTTP    = "http"
	CategoryRuntime = "runtime"
	CategoryStore   = "store"
	CategoryServer  = "server"
)

const (
	EnvGlobalLevel  = "LOG_LEVEL"
	EnvHTTPLevel    = "LOG_PLAYER_DATA_HTTP"
	EnvRuntimeLevel = "LOG_PLAYER_DATA_RUNTIME"
	EnvStoreLevel   = "LOG_PLAYER_DATA_STORE"
	EnvServerLevel  = "LOG_PLAYER_DATA_SERVER"
)

const (
	FieldCategory       = "category"
	FieldError          = "error"
	FieldIdentityKind   = "identity_kind"
	FieldLocalProfileID = "local_profile_id"
	FieldOperation      = "operation"
)

const (
	levelOff slog.Level = slog.LevelError + 1

	ServiceName                 = observability.ServiceNamePlayerData
	fileSegmentMaxBytes   int64 = 32 * 1024 * 1024
	fileRetentionMaxBytes int64 = 512 * 1024 * 1024
)

var (
	level        = new(slog.LevelVar)
	httpLevel    = new(slog.LevelVar)
	runtimeLevel = new(slog.LevelVar)
	storeLevel   = new(slog.LevelVar)
	serverLevel  = new(slog.LevelVar)
	logRuntime   *servicelog.Runtime
	eventEmitter *observability.Emitter
	identity     servicelog.ServiceIdentity
	lastStatus   servicelog.Status
)

var (
	HTTP    = newCategoryLogger(CategoryHTTP, httpLevel)
	Runtime = newCategoryLogger(CategoryRuntime, runtimeLevel)
	Store   = newCategoryLogger(CategoryStore, storeLevel)
	Server  = newCategoryLogger(CategoryServer, serverLevel)
)

type CategoryLogger struct {
	name  string
	level *slog.LevelVar
}

// Emit is the narrow semantic entry point for player-data-owned events.
// Legacy category methods remain available only for service migrations that
// have not yet moved to canonical events.
func Emit(request observability.Request) observability.Result {
	if eventEmitter == nil {
		eventEmitter = fallbackEmitter()
	}
	return eventEmitter.Emit(request)
}

func newCategoryLogger(name string, level *slog.LevelVar) CategoryLogger {
	return CategoryLogger{name: name, level: level}
}

func (logger CategoryLogger) Debug(message string, args ...any) {
	logger.log(slog.LevelDebug, message, args...)
}

func (logger CategoryLogger) Info(message string, args ...any) {
	logger.log(slog.LevelInfo, message, args...)
}

func (logger CategoryLogger) Warn(message string, args ...any) {
	logger.log(slog.LevelWarn, message, args...)
}

func (logger CategoryLogger) Error(message string, err error, args ...any) {
	logger.log(slog.LevelError, message, append(args, FieldError, err)...)
}

func (logger CategoryLogger) log(messageLevel slog.Level, message string, args ...any) {
	if logger.level == nil || messageLevel < logger.level.Level() {
		return
	}
	emitLegacy(logger.name, messageLevel, message, args...)
}

func init() {
	level.Set(slog.LevelWarn)
	configureCategoryLevels(slog.LevelWarn)
	eventEmitter = fallbackEmitter()
}

func Configure(configuredLevel string) {
	defaultLevel := parseLevel(configuredLevel)
	level.Set(defaultLevel)
	configureCategoryLevels(defaultLevel)
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

func ConfigureFileOutput(baseDir, prefix string) (string, error) {
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

func Debug(message string, args ...any) { emit(slog.LevelDebug, message, args...) }
func Info(message string, args ...any)  { emit(slog.LevelInfo, message, args...) }
func Warn(message string, args ...any)  { emit(slog.LevelWarn, message, args...) }

func Error(message string, err error, args ...any) {
	emit(slog.LevelError, message, append(args, FieldError, err)...)
}

func emit(messageLevel slog.Level, message string, args ...any) {
	if messageLevel >= level.Level() {
		emitLegacy(CategoryServer, messageLevel, message, args...)
	}
}

func emitLegacy(category string, messageLevel slog.Level, message string, args ...any) {
	if eventEmitter == nil {
		eventEmitter = fallbackEmitter()
	}
	eventEmitter.EmitLegacyArgs(observability.LegacyRequest{
		Level: observability.Level(levelName(messageLevel)), Category: category, Message: message,
	}, args...)
}

func levelName(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "error"
	case level >= slog.LevelWarn:
		return "warn"
	case level >= slog.LevelInfo:
		return "info"
	default:
		return "debug"
	}
}

func newEmitter(sink observability.Sink, configuredIdentity servicelog.ServiceIdentity) (*observability.Emitter, error) {
	return observability.New(observability.Config{
		Service: observability.ServiceKeyPlayerData, Environment: configuredIdentity.Environment,
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
		Service: observability.ServiceKeyPlayerData, Environment: "unconfigured", BuildVersion: "unknown",
		ServiceInstanceID: "00000000-0000-4000-8000-000000000000", PID: os.Getpid(), Sink: stderrSink{}, WarningWriter: os.Stderr,
	})
	return emitter
}

func validateIdentity(value servicelog.ServiceIdentity) error {
	if strings.TrimSpace(value.Name) == "" || strings.TrimSpace(value.Version) == "" || strings.TrimSpace(value.Environment) == "" || strings.TrimSpace(value.InstanceID) == "" {
		return errors.New("player-data logging: service name, build version, environment, and instance ID are required")
	}
	return nil
}

func fileRuntimeConfig(configuredIdentity servicelog.ServiceIdentity, baseDir, prefix string) servicelog.Config {
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

func activeFilePath(baseDir, prefix string) string {
	return filepath.Join(baseDir, prefix+".jsonl.open")
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

func configureCategoryLevels(defaultLevel slog.Level) {
	httpLevel.Set(parseLevelOrDefault(os.Getenv(EnvHTTPLevel), defaultLevel))
	runtimeLevel.Set(parseLevelOrDefault(os.Getenv(EnvRuntimeLevel), defaultLevel))
	storeLevel.Set(parseLevelOrDefault(os.Getenv(EnvStoreLevel), defaultLevel))
	serverLevel.Set(parseLevelOrDefault(os.Getenv(EnvServerLevel), defaultLevel))
}

func parseLevelOrDefault(configuredLevel string, defaultLevel slog.Level) slog.Level {
	if strings.TrimSpace(configuredLevel) == "" {
		return defaultLevel
	}
	return parseLevel(configuredLevel)
}
