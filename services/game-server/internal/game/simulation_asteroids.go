package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/motion"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func (game *Game) stepAsteroidSpawning(delta float64) {
	if game.worldSimulationOptions.CanSpawnAsteroids() && game.hasCameraViews() {
		game.asteroidSpawnElapsed += delta
		if game.asteroidSpawnElapsed >= constants.AsteroidSpawnInterval {
			game.asteroidSpawnElapsed = 0
			for _, cameraView := range game.cameraViews {
				game.spawnAsteroidBatch(cameraView)
			}
		}
	} else if !game.hasCameraViews() {
		game.asteroidSpawnElapsed = 0
	}
}

func (game *Game) stepAsteroids(delta float64, bounds space.Bounds) {
	for id, asteroid := range game.state.Asteroids {
		if game.worldSimulationOptions.AsteroidsCanMove() {
			motion.AdvanceAsteroid(asteroid, delta, bounds)
		}
		if asteroid.ReadyForRemoval() {
			delete(game.state.Asteroids, id)
			continue
		}
		if game.isAsteroidFarFromAllCameras(asteroid) {
			delete(game.state.Asteroids, id)
		}
	}
}
