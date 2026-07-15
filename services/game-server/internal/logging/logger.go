package logging

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/observability"
	"github.com/Lokee86/space-rocks/shared/go/servicelog"
)

const (
	CategoryGame    = "game"
	CategoryNetwork = "network"
	CategoryRooms   = "rooms"
	CategoryServer  = "server"
)

const (
	EnvGameLevel    = "LOG_GAME"
	EnvGlobalLevel  = "LOG_LEVEL"
	EnvNetworkLevel = "LOG_NETWORK"
	EnvRoomsLevel   = "LOG_ROOMS"
	EnvServerLevel  = "LOG_SERVER"
)

const (
	FieldCategory   = "category"
	FieldError      = "error"
	FieldPacketType = "packet_type"
	FieldPlayerID   = "player_id"
	FieldRemoteAddr = "remote_addr"
	FieldRoomID     = "room_id"
)

const (
	levelOff slog.Level = slog.LevelError + 1

	ServiceName                 = "game-server"
	fileSegmentMaxBytes   int64 = 64 * 1024 * 1024
	fileRetentionMaxBytes int64 = 1 * 1024 * 1024 * 1024
)

var (
	rootLevel    = new(slog.LevelVar)
	gameLevel    = new(slog.LevelVar)
	networkLevel = new(slog.LevelVar)
	roomsLevel   = new(slog.LevelVar)
	serverLevel  = new(slog.LevelVar)
	logRuntime   *servicelog.Runtime
	identity     servicelog.ServiceIdentity
	lastStatus   servicelog.Status
)

var (
	Game    = newCategoryLogger(CategoryGame, gameLevel)
	Network = newCategoryLogger(CategoryNetwork, networkLevel)
	Rooms   = newCategoryLogger(CategoryRooms, roomsLevel)
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
	rootLevel.Set(slog.LevelWarn)
	configureCategoryLevels(slog.LevelWarn)
	rebuildLoggers()
}

func Configure(configuredLevel string) {
	defaultLevel := parseLevel(configuredLevel)
	rootLevel.Set(defaultLevel)
	configureCategoryLevels(defaultLevel)
	rebuildLoggers()
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
	rebuildLoggers()
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
	rebuildLoggers()
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
	if messageLevel >= rootLevel.Level() {
		runtimeLogger().Log(context.Background(), messageLevel, message, args...)
	}
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
	gameLevel.Set(parseLevelOrDefault(os.Getenv(EnvGameLevel), defaultLevel))
	networkLevel.Set(parseLevelOrDefault(os.Getenv(EnvNetworkLevel), defaultLevel))
	roomsLevel.Set(parseLevelOrDefault(os.Getenv(EnvRoomsLevel), defaultLevel))
	serverLevel.Set(parseLevelOrDefault(os.Getenv(EnvServerLevel), defaultLevel))
}

func parseLevelOrDefault(configuredLevel string, defaultLevel slog.Level) slog.Level {
	if strings.TrimSpace(configuredLevel) == "" {
		return defaultLevel
	}
	return parseLevel(configuredLevel)
}

func rebuildLoggers() {
	Game = newCategoryLogger(CategoryGame, gameLevel)
	Network = newCategoryLogger(CategoryNetwork, networkLevel)
	Rooms = newCategoryLogger(CategoryRooms, roomsLevel)
	Server = newCategoryLogger(CategoryServer, serverLevel)
}

func runtimeLogger() *slog.Logger {
	if logRuntime != nil {
		return logRuntime.Logger()
	}
	return slog.New(slog.NewTextHandler(os.Stderr, nil))
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
