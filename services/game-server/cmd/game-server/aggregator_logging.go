package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging/aggregatorclient"
)

func configureAggregatorLoggingFromEnv() (bool, error) {
	config, err := aggregatorclient.ConfigFromEnv()
	if err != nil {
		return false, fmt.Errorf("invalid aggregator logging configuration: %w", err)
	}
	if err := logging.ConfigureAggregatorOutput(config); err != nil {
		return false, fmt.Errorf("configure aggregator logging: %w", err)
	}
	return config.Enabled, nil
}

func closeAggregatorLogging(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return logging.CloseAggregatorOutput(ctx)
}
