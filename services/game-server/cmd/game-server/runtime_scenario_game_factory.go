package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

const runtimeScenarioSeedEnv = "SPACE_ROCKS_RUNTIME_SCENARIO_SEED"

func runtimeScenarioGameFactoryFromEnv() (rooms.GameFactory, error) {
	rawSeed := strings.TrimSpace(os.Getenv(runtimeScenarioSeedEnv))
	if rawSeed == "" {
		return nil, nil
	}
	seed, err := strconv.ParseInt(rawSeed, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s must be a signed 64-bit integer: %w", runtimeScenarioSeedEnv, err)
	}
	return func() *game.Game {
		return game.NewWithSeed(seed)
	}, nil
}
