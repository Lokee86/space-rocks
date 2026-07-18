package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/motion"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

func (game *Game) stepAsteroidSpawning(delta float64) {
	runtime := game.encounterSpawn()
	if !game.hasCameraViews() {
		if err := runtime.ResetProgress(encounterspawn.ProfilePlayercentricAsteroidsV1); err != nil {
			panic(fmt.Errorf("failed to reset baseline encounter spawn progress: %w", err))
		}
		return
	}

	opportunities, err := runtime.Step(delta, !game.worldSimulationOptions.CanSpawnAsteroids())
	if err != nil {
		panic(fmt.Errorf("failed to step encounter spawn profiles: %w", err))
	}
	for _, opportunity := range opportunities {
		if opportunity.ProfileID != encounterspawn.ProfilePlayercentricAsteroidsV1 {
			panic(fmt.Errorf("unsupported encounter spawn profile %q", opportunity.ProfileID))
		}
		snapshot, _ := runtime.Snapshot(opportunity.ProfileID)
		for _, cameraView := range game.sortedCameraViews() {
			game.spawnAsteroidBatchForProfile(cameraView, opportunity.ProfileID, opportunity.BatchSize, snapshot.Config.RetryCap)
		}
	}
}

func (game *Game) stepAsteroids(delta float64, bounds space.Bounds) {
	for id, asteroid := range game.entities.Asteroids {
		game.ensureAsteroidLifecycleRegistered(asteroid)
		if game.worldSimulationOptions.AsteroidsCanMove() {
			motion.AdvanceAsteroid(asteroid, delta, bounds)
		}
		if asteroid.ReadyForRemoval() {
			game.removeAsteroidAuthoritatively(id)
			continue
		}
		game.evaluateAsteroidLifecycle(asteroid)
	}
}
