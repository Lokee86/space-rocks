package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	runtimepkg "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

func (target *Control) RandomUnitVector() physics.Vector2 {
	return target.game.spawner.RandomUnitVector()
}

func (target *Control) NextBulletID() string {
	return target.game.spawner.NextBulletID()
}

func (target *Control) AddBullet(bullet *runtimepkg.Bullet) bool {
	if bullet == nil {
		return false
	}
	target.game.entities.Projectiles[bullet.ID] = bullet
	return true
}

func (target *Control) SpawnBullet(ownerID string, position physics.Vector2, direction physics.Vector2) (*runtimepkg.Bullet, bool) {
	return target.game.spawnDebugBullet(ownerID, position, direction)
}

func (target *Control) SpawnPickup(pickupType pickups.PickupType, position physics.Vector2) (*pickups.Pickup, bool, error) {
	return target.game.SpawnPickup(pickupType, position)
}

func (target *Control) RandomAsteroidSpeed() float64 {
	return target.game.spawner.RandomAsteroidSpeed()
}

func (target *Control) ApplyAsteroidSpawnPlan(plan spawning.AsteroidSpawnPlan) *runtimepkg.Asteroid {
	return target.game.applyAsteroidSpawn(plan)
}
