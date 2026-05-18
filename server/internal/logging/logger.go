package logging

import (
	"log/slog"
	"os"
	"strings"
)

const (
	FieldError      = "error"
	FieldPacketType = "packet_type"
	FieldPlayerID   = "player_id"
	FieldRemoteAddr = "remote_addr"
	FieldRoomID     = "room_id"
)

var level = new(slog.LevelVar)

func init() {
	level.Set(slog.LevelInfo)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))
}

func Configure(configuredLevel string) {
	level.Set(parseLevel(configuredLevel))
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
	default:
		return slog.LevelInfo
	}
}
