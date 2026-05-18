package logging

import (
	"log/slog"
	"os"
	"strings"
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

const levelOff slog.Level = slog.LevelError + 1

var (
	level        = new(slog.LevelVar)
	gameLevel    = new(slog.LevelVar)
	networkLevel = new(slog.LevelVar)
	roomsLevel   = new(slog.LevelVar)
	serverLevel  = new(slog.LevelVar)
)

var (
	Game    = newCategoryLogger(CategoryGame, gameLevel)
	Network = newCategoryLogger(CategoryNetwork, networkLevel)
	Rooms   = newCategoryLogger(CategoryRooms, roomsLevel)
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
