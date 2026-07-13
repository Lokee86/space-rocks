package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging/aggregatorclient"
)

const aggregatorReplacementDrainTimeout = 5 * time.Second

var aggregatorOutputState struct {
	sync.Mutex
	sink *aggregatorclient.Sink
}

// ConfigureAggregatorOutput replaces the optional aggregator logging sink.
func ConfigureAggregatorOutput(config aggregatorclient.Config) error {
	newSink, err := aggregatorclient.New(config)
	if err != nil {
		return err
	}

	aggregatorOutputState.Lock()
	old := aggregatorOutputState.sink
	aggregatorOutputState.sink = newSink
	aggregatorOutputState.Unlock()
	rebuildLoggers()

	if old == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), aggregatorReplacementDrainTimeout)
	defer cancel()
	if err := old.Close(ctx); err != nil {
		return fmt.Errorf("drain replaced aggregator sink: %w", err)
	}
	return nil
}

// CloseAggregatorOutput detaches and drains the configured aggregator sink.
func CloseAggregatorOutput(ctx context.Context) error {
	aggregatorOutputState.Lock()
	old := aggregatorOutputState.sink
	aggregatorOutputState.sink = nil
	aggregatorOutputState.Unlock()
	rebuildLoggers()
	if old == nil {
		return nil
	}
	return old.Close(ctx)
}

// AggregatorStats returns a snapshot of the active sink counters.
func AggregatorStats() aggregatorclient.Stats {
	aggregatorOutputState.Lock()
	defer aggregatorOutputState.Unlock()
	if aggregatorOutputState.sink == nil {
		return aggregatorclient.Stats{}
	}
	return aggregatorOutputState.sink.Stats()
}
