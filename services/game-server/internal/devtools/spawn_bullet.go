package devtools

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func debugBulletSpawnPosition(request SpawnEntityRequest) physics.Vector2 {
	return space.NormalizePosition(request.Position())
}

func debugBulletDirection(target *game.Game, request SpawnEntityRequest) physics.Vector2 {
	return request.DirectionOr(target.DevtoolsRandomUnitVector())
}

func debugBulletRotation(direction physics.Vector2) float64 {
	return math.Atan2(direction.X, -direction.Y)
}

func buildDebugBullet(target *game.Game, bulletID string, ownerID string, request SpawnEntityRequest) *entities.Bullet {
	if target == nil || bulletID == "" || ownerID == "" {
		return nil
	}

	position := debugBulletSpawnPosition(request)
	direction := debugBulletDirection(target, request)
	velocity := direction.Multiply(constants.BulletSpeed)
	rotation := debugBulletRotation(direction)
	return entities.NewBullet(bulletID, ownerID, position, rotation, velocity, constants.BulletLifetime)
}

func applyDebugSpawnBullet(target *game.Game, ownerID string, request SpawnEntityRequest) (*entities.Bullet, bool) {
	if target == nil || ownerID == "" {
		return nil, false
	}

	bulletID := target.DevtoolsNextBulletID()
	bullet := buildDebugBullet(target, bulletID, ownerID, request)
	if bullet == nil {
		return nil, false
	}
	if !target.DevtoolsAddBullet(bullet) {
		return nil, false
	}

	return bullet, true
}
