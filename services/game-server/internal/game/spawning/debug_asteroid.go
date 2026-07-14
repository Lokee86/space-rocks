package spawning

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/asteroids"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func (spawner *Spawner) PlanDebugAsteroidSpawn(position physics.Vector2, requestedDirection physics.Vector2, hasRequestedDirection bool) AsteroidSpawnPlan {
	fallbackDirection := spawner.RandomUnitVector()
	direction := requestedDirection
	if !hasRequestedDirection || direction.Length() == 0 {
		direction = fallbackDirection
	} else {
		direction = direction.Normalized()
	}

	debugVariants := asteroids.DebugSpawnVariants()
	variantIndex := 0
	if len(debugVariants) > 0 {
		totalWeight := 0.0
		for _, variant := range debugVariants {
			totalWeight += variant.DebugSpawnWeight
		}
		if totalWeight > 0.0 {
			threshold := spawner.rngSource.Float64() * totalWeight
			var cumulativeWeight float64
			variantIndex = debugVariants[len(debugVariants)-1].Index
			for _, variant := range debugVariants {
				cumulativeWeight += variant.DebugSpawnWeight
				if threshold < cumulativeWeight {
					variantIndex = variant.Index
					break
				}
			}
		}
	}

	return AsteroidSpawnPlan{
		EntityType: SpawnEntityTypeAsteroid,
		Reason:     SpawnReasonDebugAsteroid,
		Position:   position,
		Velocity:   direction.Multiply(spawner.RandomAsteroidSpeed()),
		Size:       spawner.rngSource.Intn(4) + 1,
		Variant:    variantIndex,
	}
}
