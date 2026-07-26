package game

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/motion"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

func (game *Game) stepAsteroidSpawning(delta float64) {
	if !game.worldSimulationOptions.CanSpawnAsteroids() || !game.hasCameraViews() {
		if !game.hasCameraViews() {
			game.asteroidSpawnElapsed = 0
			game.asteroidSpawnCameraCursor = 0
		}
		return
	}

	game.asteroidSpawnElapsed += delta
	if game.asteroidSpawnElapsed < constants.AsteroidSpawnInterval {
		return
	}

	cameraViews := game.sortedTimedSpawnCameraViews()
	if len(cameraViews) == 0 {
		game.asteroidSpawnElapsed = 0
		game.asteroidSpawnCameraCursor = 0
		return
	}

	remainingCapacity := timedAsteroidPopulationLimit(len(cameraViews)) - len(game.entities.Asteroids)
	if remainingCapacity <= 0 {
		game.asteroidSpawnElapsed = constants.AsteroidSpawnInterval
		return
	}

	game.asteroidSpawnElapsed = 0
	spawnBudget := constants.AsteroidSpawnBatchSize
	if len(cameraViews) > spawnBudget {
		spawnBudget = len(cameraViews)
	}
	if remainingCapacity < spawnBudget {
		spawnBudget = remainingCapacity
	}

	for offset := 0; offset < spawnBudget; offset++ {
		cameraIndex := (game.asteroidSpawnCameraCursor + offset) % len(cameraViews)
		game.spawnAsteroid(cameraViews[cameraIndex])
	}
	game.asteroidSpawnCameraCursor = (game.asteroidSpawnCameraCursor + spawnBudget) % len(cameraViews)
}

func timedAsteroidPopulationLimit(cameraCount int) int {
	if cameraCount <= 0 {
		return 0
	}
	return constants.AsteroidTimedSpawnBaseLimit + cameraCount*constants.AsteroidTimedSpawnPerCameraLimit
}

func (game *Game) sortedTimedSpawnCameraViews() []*runtime.CameraView {
	cameraIDs := make([]string, 0, len(game.cameraViews))
	for cameraID, cameraView := range game.cameraViews {
		if cameraView == nil {
			continue
		}
		cameraIDs = append(cameraIDs, cameraID)
	}
	sort.Strings(cameraIDs)

	cameraViews := make([]*runtime.CameraView, 0, len(cameraIDs))
	for _, cameraID := range cameraIDs {
		cameraViews = append(cameraViews, game.cameraViews[cameraID])
	}
	return cameraViews
}

func (game *Game) stepAsteroids(delta float64, bounds space.Bounds) {
	for id, asteroid := range game.entities.Asteroids {
		if game.worldSimulationOptions.AsteroidsCanMove() {
			motion.AdvanceAsteroid(asteroid, delta, bounds)
		}
		if asteroid.ReadyForRemoval() {
			delete(game.entities.Asteroids, id)
			continue
		}
		if game.isAsteroidFarFromAllCameras(asteroid) {
			delete(game.entities.Asteroids, id)
		}
	}
}
