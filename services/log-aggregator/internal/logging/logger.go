package logging

import (
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/config"
	"github.com/Lokee86/space-rocks/services/log-aggregator/internal/serviceidentity"
)

func New(cfg config.Config, consoleWriter, fileWriter io.Writer) (*slog.Logger, error) {
	level, err := parseLevel(cfg.LogLevel)
	if err != nil {
		return nil, err
	}
	options := &slog.HandlerOptions{Level: level}
	handlers := make([]slog.Handler, 0, 2)
	if cfg.ConsoleLogging {
		if consoleWriter == nil {
			return nil, fmt.Errorf("logging: console sink is enabled but has no writer")
		}
		handlers = append(handlers, slog.NewJSONHandler(consoleWriter, options))
	}
	if cfg.FileLogging {
		if fileWriter == nil {
			return nil, fmt.Errorf("logging: file sink is enabled but has no writer")
		}
		handlers = append(handlers, slog.NewJSONHandler(fileWriter, options))
	}
	var handler slog.Handler
	switch len(handlers) {
	case 0:
		handler = slog.NewJSONHandler(io.Discard, options)
	case 1:
		handler = handlers[0]
	default:
		handler = newFanout(handlers...)
	}
	return slog.New(handler).With("service", serviceidentity.ServiceName, "service_instance_id", cfg.ServiceInstanceID, "environment", cfg.Environment, "build_version", cfg.BuildVersion), nil
}

func NewLogger(cfg config.Config, consoleWriter, fileWriter io.Writer) (*slog.Logger, error) {
	return New(cfg, consoleWriter, fileWriter)
}

func parseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(value) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("logging: invalid minimum log level %q", value)
	}
}
