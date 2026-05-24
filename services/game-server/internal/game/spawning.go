package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) spawnBullet(ship *entities.Ship) {
	bullet := game.spawner.BuildBullet(ship)
	game.state.Projectiles[bullet.ID] = bullet
}

func (game *Game) spawnAsteroidBatch(view *entities.CameraView) {
	for range constants.AsteroidSpawnBatchSize {
		game.spawnAsteroid(view)
	}
}

func (game *Game) spawnAsteroid(view *entities.CameraView) {
	targetPosition := view.Position()
	spawn := game.randomAsteroidSpawnPosition(view)
	spawn = space.NormalizePosition(spawn)
	plan := game.spawner.PlanTimedAsteroidSpawn(spawn, targetPosition)
	game.applyAsteroidSpawn(plan)
}

func (game *Game) applyAsteroidSpawn(plan spawning.AsteroidSpawnPlan) *entities.Asteroid {
	asteroidID := game.spawner.NextAsteroidID(game.state.Asteroids)
	asteroid := entities.NewAsteroid(asteroidID, plan.Position, plan.Velocity, plan.Size, plan.Variant)
	game.state.Asteroids[asteroidID] = asteroid
	return asteroid
}

func (game *Game) spawnAsteroidFragments(asteroid *entities.Asteroid) {
	fragmentSize := asteroid.FragmentSize()
	if fragmentSize <= 0 {
		return
	}

	position := asteroid.Position()
	logging.Game.Debug("asteroid split",
		"asteroid_id", asteroid.ID,
		"source_size", asteroid.Size,
		"fragment_size", fragmentSize,
		"x", position.X,
		"y", position.Y,
	)
	plans := game.spawner.PlanAsteroidFragmentSpawns(asteroid)
	for _, plan := range plans {
		game.applyAsteroidSpawn(plan)
	}
}
