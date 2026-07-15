package logging

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/player-data/observability"
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

	ServiceName                 = "player-data"
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
	runtimeLogger().With(slog.String(FieldCategory, logger.name)).Log(context.Background(), messageLevel, message, args...)
}

func init() {
	level.Set(slog.LevelWarn)
	configureCategoryLevels(slog.LevelWarn)
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

func replaceRuntime(runtime *servicelog.Runtime, configuredIdentity servicelog.ServiceIdentity) error {
	oldRuntime := logRuntime
	logRuntime = runtime
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
		runtimeLogger().Log(context.Background(), messageLevel, message, args...)
	}
}

func runtimeLogger() *slog.Logger {
	if logRuntime != nil {
		return logRuntime.Logger()
	}
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
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
