package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

func (game *Game) DevtoolsRandomUnitVector() physics.Vector2 {
	return game.spawner.RandomUnitVector()
}

func (game *Game) DevtoolsNextBulletID() string {
	return game.spawner.NextBulletID()
}

func (game *Game) DevtoolsAddBullet(bullet *runtime.Bullet) bool {
	if bullet == nil {
		return false
	}
	game.state.Projectiles[bullet.ID] = bullet
	return true
}

func (game *Game) DevtoolsSpawnBullet(ownerID string, position physics.Vector2, direction physics.Vector2) (*runtime.Bullet, bool) {
	return game.spawnDebugBullet(ownerID, position, direction)
}

func (game *Game) DevtoolsRandomAsteroidSpeed() float64 {
	return game.spawner.RandomAsteroidSpeed()
}

func (game *Game) DevtoolsApplyAsteroidSpawnPlan(plan spawning.AsteroidSpawnPlan) *runtime.Asteroid {
	return game.applyAsteroidSpawn(plan)
}
