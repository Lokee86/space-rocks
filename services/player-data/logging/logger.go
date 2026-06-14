package logging

import (
	"log/slog"
	"os"
	"strings"
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

const levelOff slog.Level = slog.LevelError + 1

var (
	level        = new(slog.LevelVar)
	httpLevel    = new(slog.LevelVar)
	runtimeLevel = new(slog.LevelVar)
	storeLevel   = new(slog.LevelVar)
	serverLevel  = new(slog.LevelVar)
)

var (
	HTTP    = newCategoryLogger(CategoryHTTP, httpLevel)
	Runtime = newCategoryLogger(CategoryRuntime, runtimeLevel)
	Store   = newCategoryLogger(CategoryStore, storeLevel)
	Server  = newCategoryLogger(CategoryServer, serverLevel)
)

type CategoryLogger struct {
	name   string
	level  *slog.LevelVar
	logger *slog.Logger
}

func newCategoryLogger(name string, level *slog.LevelVar) CategoryLogger {
	return CategoryLogger{
		name:  name,
		level: level,
		logger: slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})),
	}
}

func (logger CategoryLogger) Debug(message string, args ...any) {
	logger.logger.Debug(message, logger.args(args)...)
}

func (logger CategoryLogger) Info(message string, args ...any) {
	logger.logger.Info(message, logger.args(args)...)
}

func (logger CategoryLogger) Warn(message string, args ...any) {
	logger.logger.Warn(message, logger.args(args)...)
}

func (logger CategoryLogger) Error(message string, err error, args ...any) {
	args = append(args, FieldError, err)
	logger.logger.Error(message, logger.args(args)...)
}

func (logger CategoryLogger) args(args []any) []any {
	return append([]any{FieldCategory, logger.name}, args...)
}

func init() {
	level.Set(slog.LevelWarn)
	configureCategoryLevels(slog.LevelWarn)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}

func Configure(configuredLevel string) {
	defaultLevel := parseLevel(configuredLevel)
	level.Set(defaultLevel)
	configureCategoryLevels(defaultLevel)
}

func Debug(message string, args ...any) {
	slog.Debug(message, args...)
}

func Info(message string, args ...any) {
	slog.Info(message, args...)
}

func Warn(message string, args ...any) {
	slog.Warn(message, args...)
}

func Error(message string, err error, args ...any) {
	args = append(args, FieldError, err)
	slog.Error(message, args...)
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
