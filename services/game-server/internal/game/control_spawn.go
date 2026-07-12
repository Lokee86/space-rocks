package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	runtimepkg "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

func (target *Control) RandomUnitVector() physics.Vector2 {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.spawner.RandomUnitVector()
}

func (target *Control) NextBulletID() string {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.spawner.NextBulletID()
}

func (target *Control) AddBullet(bullet *runtimepkg.Bullet) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	if bullet == nil {
		return false
	}
	target.game.entities.Projectiles[bullet.ID] = bullet
	target.game.publishPresentationFrameLocked()
	return true
}

func (target *Control) SpawnBullet(ownerID string, position physics.Vector2, direction physics.Vector2) (*runtimepkg.Bullet, bool) {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	bullet, ok := target.game.spawnDebugBullet(ownerID, position, direction)
	if ok {
		target.game.publishPresentationFrameLocked()
	}
	return bullet, ok
}

func (target *Control) SpawnPickup(pickupType pickups.PickupType, position physics.Vector2) (*pickups.Pickup, bool, error) {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	pickup, ok, err := target.game.spawnPickupLocked(pickupType, position)
	if ok {
		target.game.publishPresentationFrameLocked()
	}
	return pickup, ok, err
}

func (target *Control) RandomAsteroidSpeed() float64 {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.spawner.RandomAsteroidSpeed()
}

func (target *Control) ApplyAsteroidSpawnPlan(plan spawning.AsteroidSpawnPlan) *runtimepkg.Asteroid {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	asteroid := target.game.applyAsteroidSpawn(plan)
	if asteroid != nil {
		target.game.publishPresentationFrameLocked()
	}
	return asteroid
}
