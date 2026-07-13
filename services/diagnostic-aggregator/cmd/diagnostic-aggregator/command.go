package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/config"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/logging"
	"github.com/Lokee86/space-rocks/services/diagnostic-aggregator/internal/serviceidentity"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitConfig  = 2
)

type AppRunner interface{ Run(context.Context) error }
type ConfigLoader func() (config.Config, error)
type ServiceLogOpener func(config.Config) (io.WriteCloser, error)
type AppFactory func(config.Config, *slog.Logger) (AppRunner, error)

func run(ctx context.Context, stderr io.Writer, loadConfig ConfigLoader, openServiceLog ServiceLogOpener, buildApp AppFactory) int {
	if stderr == nil {
		stderr = io.Discard
	}
	if loadConfig == nil {
		bootstrapError(stderr, "configuration_invalid", "configuration is invalid")
		return exitConfig
	}
	cfg, err := loadConfig()
	if err != nil {
		bootstrapError(stderr, "configuration_invalid", "configuration is invalid")
		return exitConfig
	}
	var serviceLog io.WriteCloser
	degraded := false
	if cfg.FileLogging {
		if openServiceLog != nil {
			serviceLog, err = openServiceLog(cfg)
		}
		if err != nil || serviceLog == nil {
			if serviceLog != nil {
				_ = closeServiceLog(stderr, serviceLog)
			}
			serviceLog = nil
			degraded = true
		}
	}
	loggerConfig := cfg
	if degraded {
		loggerConfig.FileLogging = false
	}
	logger, err := logging.New(loggerConfig, stderr, serviceLog)
	if err != nil {
		_ = closeServiceLog(stderr, serviceLog)
		bootstrapError(stderr, "configuration_invalid", "logging configuration is invalid")
		return exitConfig
	}
	if degraded {
		logger.Warn("diagnostic aggregator environment degraded", "event", "environment_degraded", "error_code", "service_log_file_unavailable")
	}
	if buildApp == nil {
		logger.Error("diagnostic aggregator startup failed", "event", "service_startup_failed", "error_code", "runtime_build_failed")
		_ = closeServiceLog(stderr, serviceLog)
		return exitFailure
	}
	app, err := buildApp(cfg, logger)
	if err != nil {
		logger.Error("diagnostic aggregator startup failed", "event", "service_startup_failed", "error_code", "runtime_build_failed")
		_ = closeServiceLog(stderr, serviceLog)
		return exitFailure
	}
	runErr := app.Run(ctx)
	closeErr := closeServiceLog(stderr, serviceLog)
	if runErr != nil || closeErr != nil {
		return exitFailure
	}
	return exitSuccess
}

func closeServiceLog(stderr io.Writer, writer io.WriteCloser) error {
	if writer == nil {
		return nil
	}
	err := writer.Close()
	if err != nil {
		bootstrapError(stderr, "service_log_close_failed", "service log close failed")
	}
	return err
}

func bootstrapError(stderr io.Writer, code, message string) {
	if stderr == nil {
		stderr = io.Discard
	}
	_ = json.NewEncoder(stderr).Encode(map[string]string{"event": "service_startup_failed", "service": serviceidentity.ServiceName, "error_code": code, "message": message})
}
