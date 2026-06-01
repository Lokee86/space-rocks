package devtools

import (
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

func applyDebugSpawnBullet(target *game.Game, ownerID string, request SpawnEntityRequest) (*entities.Bullet, bool) {
	if target == nil || ownerID == "" {
		return nil, false
	}
	position := debugBulletSpawnPosition(request)
	direction := debugBulletDirection(target, request)
	return target.DevtoolsSpawnBullet(ownerID, position, direction)
}
