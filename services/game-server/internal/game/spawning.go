package game

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) spawnBullet(ship *runtime.Ship) {
	bullet := game.spawner.BuildBullet(ship)
	game.state.Projectiles[bullet.ID] = bullet
}

func debugBulletRotation(direction physics.Vector2) float64 {
	return math.Atan2(direction.X, -direction.Y)
}

func (game *Game) spawnDebugBullet(ownerID string, position physics.Vector2, direction physics.Vector2) (*runtime.Bullet, bool) {
	if ownerID == "" {
		return nil, false
	}
	normalizedDirection := direction.Normalized()
	if normalizedDirection.Length() == 0 {
		return nil, false
	}
	spawnPosition := space.NormalizePosition(position)
	velocity := normalizedDirection.Multiply(constants.BulletSpeed)
	rotation := debugBulletRotation(normalizedDirection)
	bulletID := game.spawner.NextBulletID()
	bullet := runtime.NewBullet(bulletID, ownerID, spawnPosition, rotation, velocity, constants.BulletLifetime)
	game.state.Projectiles[bullet.ID] = bullet
	return bullet, true
}

func (game *Game) spawnAsteroidBatch(view *runtime.CameraView) {
	for range constants.AsteroidSpawnBatchSize {
		game.spawnAsteroid(view)
	}
}

func (game *Game) spawnAsteroid(view *runtime.CameraView) {
	targetPosition := view.Position()
	spawn := game.randomAsteroidSpawnPosition(view)
	spawn = space.NormalizePosition(spawn)
	plan := game.spawner.PlanTimedAsteroidSpawn(spawn, targetPosition)
	game.applyAsteroidSpawn(plan)
}

func (game *Game) applyAsteroidSpawn(plan spawning.AsteroidSpawnPlan) *runtime.Asteroid {
	asteroidID := game.spawner.NextAsteroidID(game.state.Asteroids)
	asteroid := runtime.NewAsteroid(asteroidID, plan.Position, plan.Velocity, plan.Size, plan.Variant)
	game.state.Asteroids[asteroidID] = asteroid
	return asteroid
}

func (game *Game) spawnAsteroidFragments(asteroid *runtime.Asteroid) {
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
