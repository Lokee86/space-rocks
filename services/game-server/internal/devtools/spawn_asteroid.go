package devtools

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

func buildDebugAsteroidSpawnPlan(target Target, request SpawnEntityRequest) spawning.AsteroidSpawnPlan {
	normalizedPosition := space.NormalizePosition(request.Position())
	requestedDirection := physics.Vector2{X: request.DirectionX, Y: request.DirectionY}
	return target.PlanDebugAsteroidSpawn(normalizedPosition, requestedDirection, request.HasDirection)
}

func applyDebugSpawnAsteroid(target Target, request SpawnEntityRequest) (*runtime.Asteroid, spawning.AsteroidSpawnPlan, bool) {
	if target == nil {
		return nil, spawning.AsteroidSpawnPlan{}, false
	}

	plan := buildDebugAsteroidSpawnPlan(target, request)
	asteroid := target.ApplyAsteroidSpawnPlan(plan)
	if asteroid == nil {
		return nil, spawning.AsteroidSpawnPlan{}, false
	}

	return asteroid, plan, true
}
