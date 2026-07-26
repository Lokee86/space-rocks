package game

import (
	"fmt"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestStepAsteroidSpawningKeepsSingleCameraBatchSize(t *testing.T) {
	game := NewWithSeed(1)
	addTimedSpawnTestCamera(game, "player-1", 100, 100)

	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval)

	if got := len(game.entities.Asteroids); got != constants.AsteroidSpawnBatchSize {
		t.Fatalf("spawned asteroids = %d, want single-camera batch size %d", got, constants.AsteroidSpawnBatchSize)
	}
}

func TestStepAsteroidSpawningUsesGlobalBudgetAcrossCameraViews(t *testing.T) {
	game := NewWithSeed(1)
	const cameraCount = 8
	for index := 0; index < cameraCount; index++ {
		addTimedSpawnTestCamera(game, fmt.Sprintf("player-%d", index+1), float64(index*1000), 100)
	}

	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval)

	if got := len(game.entities.Asteroids); got != cameraCount {
		t.Fatalf("spawned asteroids = %d, want one global-budget spawn per camera %d", got, cameraCount)
	}
}

func TestStepAsteroidSpawningStopsAtTimedPopulationLimit(t *testing.T) {
	game := NewWithSeed(1)
	const cameraCount = 8
	for index := 0; index < cameraCount; index++ {
		addTimedSpawnTestCamera(game, fmt.Sprintf("player-%d", index+1), float64(index*1000), 100)
	}

	limit := timedAsteroidPopulationLimit(cameraCount)
	for index := 0; index < limit-2; index++ {
		asteroidID := fmt.Sprintf("asteroid-%d", index+1)
		game.entities.Asteroids[asteroidID] = &runtime.Asteroid{ID: asteroidID}
	}

	game.stepAsteroidSpawning(constants.AsteroidSpawnInterval)

	if got := len(game.entities.Asteroids); got != limit {
		t.Fatalf("asteroid population = %d, want timed population limit %d", got, limit)
	}
}

func addTimedSpawnTestCamera(game *Game, id string, x float64, y float64) {
	game.cameraViews[id] = &runtime.CameraView{
		X: x,
		Y: y,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  640,
			VisibleWorldHeight: 360,
		},
	}
}
