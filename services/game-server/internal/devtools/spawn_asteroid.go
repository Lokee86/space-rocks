package devtools

import (
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/asteroids"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

func buildDebugAsteroidSpawnPlan(target *game.Game, request SpawnEntityRequest) spawning.AsteroidSpawnPlan {
	normalizedPosition := space.NormalizePosition(request.Position())
	fallbackDirection := target.DevtoolsRandomUnitVector()
	direction := request.DirectionOr(fallbackDirection)
	speed := target.DevtoolsRandomAsteroidSpeed()
	return spawning.AsteroidSpawnPlan{
		EntityType: spawning.SpawnEntityTypeAsteroid,
		Reason:     spawning.SpawnReasonDebugAsteroid,
		Position:   normalizedPosition,
		Velocity:   direction.Multiply(speed),
		Size:       rand.Intn(4) + 1,
		Variant:    asteroids.RandomDebugSpawnVariantIndex(),
	}
}

func applyDebugSpawnAsteroid(target *game.Game, request SpawnEntityRequest) (*runtime.Asteroid, spawning.AsteroidSpawnPlan, bool) {
	if target == nil {
		return nil, spawning.AsteroidSpawnPlan{}, false
	}

	plan := buildDebugAsteroidSpawnPlan(target, request)
	asteroid := target.DevtoolsApplyAsteroidSpawnPlan(plan)
	if asteroid == nil {
		return nil, spawning.AsteroidSpawnPlan{}, false
	}

	return asteroid, plan, true
}
