package devtools

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

func debugBulletSpawnPosition(request SpawnEntityRequest) physics.Vector2 {
	return space.NormalizePosition(request.Position())
}

func debugBulletDirection(target Target, request SpawnEntityRequest) physics.Vector2 {
	return request.DirectionOr(target.RandomUnitVector())
}

func applyDebugSpawnBullet(target Target, ownerID string, request SpawnEntityRequest) (*runtime.Bullet, bool) {
	if target == nil || ownerID == "" {
		return nil, false
	}
	position := debugBulletSpawnPosition(request)
	direction := debugBulletDirection(target, request)
	return target.SpawnBullet(ownerID, position, direction)
}
