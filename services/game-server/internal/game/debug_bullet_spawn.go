package game

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/debugging"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func (game *Game) debugBulletSpawnPosition(request debugging.SpawnEntityRequest) physics.Vector2 {
	return space.NormalizePosition(request.Position())
}

func (game *Game) debugBulletDirection(request debugging.SpawnEntityRequest) physics.Vector2 {
	return request.DirectionOr(game.spawner.RandomUnitVector())
}

func debugBulletRotation(direction physics.Vector2) float64 {
	return math.Atan2(direction.X, -direction.Y)
}

func (game *Game) buildDebugBullet(bulletID string, ownerID string, request debugging.SpawnEntityRequest) *entities.Bullet {
	if bulletID == "" {
		return nil
	}
	if ownerID == "" {
		return nil
	}

	position := game.debugBulletSpawnPosition(request)
	direction := game.debugBulletDirection(request)
	velocity := direction.Multiply(constants.BulletSpeed)
	rotation := debugBulletRotation(direction)
	return entities.NewBullet(bulletID, ownerID, position, rotation, velocity, constants.BulletLifetime)
}

func (game *Game) applyDebugSpawnBullet(ownerID string, request debugging.SpawnEntityRequest) (string, bool) {
	if ownerID == "" {
		return "", false
	}

	bulletID := game.spawner.NextBulletID()
	bullet := game.buildDebugBullet(bulletID, ownerID, request)
	if bullet == nil {
		return "", false
	}

	game.state.Projectiles[bullet.ID] = bullet
	return bullet.ID, true
}
