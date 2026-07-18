package game

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterspawn"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

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
	velocity := normalizedDirection.Multiply(constants.BasicCannonProjectileSpeed)
	rotation := debugBulletRotation(normalizedDirection)
	bulletID := game.spawner.NextBulletID()
	bullet := runtime.NewBullet(bulletID, ownerID, spawnPosition, rotation, velocity, constants.BasicCannonProjectileLifetime)
	game.entities.Projectiles[bullet.ID] = bullet
	return bullet, true
}

func (game *Game) spawnAsteroidBatchForProfile(view *runtime.CameraView, profileID encounterspawn.ProfileID, batchSize int, retryCap int) int {
	spawned := 0
	for range batchSize {
		if game.spawnAsteroidForProfile(view, profileID, retryCap) {
			spawned++
		}
	}
	return spawned
}

func (game *Game) spawnAsteroidForProfile(view *runtime.CameraView, profileID encounterspawn.ProfileID, retryCap int) bool {
	spawn, ok := game.randomAsteroidSpawnPosition(view, retryCap)
	if !ok {
		return false
	}
	spawn = space.NormalizePosition(spawn)
	plan := game.spawner.PlanTimedAsteroidSpawn(spawn, view.Position())
	cost := encounterspawn.WeightedPopulation(max(plan.Size, 1))
	if !game.canAdmitEncounterSpawn(profileID, string(plan.EntityType), cost) {
		return false
	}
	game.applyAsteroidSpawnForProfile(plan, profileID)
	return true
}

func (game *Game) applyAsteroidSpawn(plan spawning.AsteroidSpawnPlan) *runtime.Asteroid {
	return game.applyAsteroidSpawnForProfile(plan, encounterspawn.ProfilePlayercentricAsteroidsV1)
}

func (game *Game) applyAsteroidSpawnForProfile(plan spawning.AsteroidSpawnPlan, profileID encounterspawn.ProfileID) *runtime.Asteroid {
	asteroidID := game.spawner.NextAsteroidID(game.entities.Asteroids)
	asteroid := runtime.NewAsteroid(asteroidID, plan.Position, plan.Velocity, plan.Size, plan.Variant)
	game.registerAsteroidLifecycleForProfile(asteroid, profileID)
	game.entities.Asteroids[asteroidID] = asteroid
	return asteroid
}

func (game *Game) spawnAsteroidFragments(asteroid *runtime.Asteroid) {
	fragmentSize := asteroid.FragmentSize()
	if fragmentSize <= 0 {
		return
	}

	profileID := encounterspawn.ProfilePlayercentricAsteroidsV1
	if entry, exists := game.encounterLifecycle().Snapshot(asteroid.ID); exists {
		profileID = encounterspawn.ProfileID(entry.Registration.Origin.ProfileID)
	}
	plans := game.spawner.PlanAsteroidFragmentSpawns(asteroid)
	for _, plan := range plans {
		game.applyAsteroidSpawnForProfile(plan, profileID)
	}
}
