package devtools

import (
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/game/asteroids"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

func buildDebugAsteroidSpawnPlan(target Target, request SpawnEntityRequest) spawning.AsteroidSpawnPlan {
	normalizedPosition := space.NormalizePosition(request.Position())
	fallbackDirection := target.RandomUnitVector()
	direction := request.DirectionOr(fallbackDirection)
	speed := target.RandomAsteroidSpeed()
	return spawning.AsteroidSpawnPlan{
		EntityType: spawning.SpawnEntityTypeAsteroid,
		Reason:     spawning.SpawnReasonDebugAsteroid,
		Position:   normalizedPosition,
		Velocity:   direction.Multiply(speed),
		Size:       rand.Intn(4) + 1,
		Variant:    asteroids.RandomDebugSpawnVariantIndex(),
	}
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
