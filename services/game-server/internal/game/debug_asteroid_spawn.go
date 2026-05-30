package game

import (
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/game/debugging"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

func (game *Game) buildDebugAsteroidSpawnPlan(request debugging.SpawnEntityRequest) spawning.AsteroidSpawnPlan {
	normalizedPosition := space.NormalizePosition(request.Position())
	fallbackDirection := game.spawner.RandomUnitVector()
	direction := request.DirectionOr(fallbackDirection)
	speed := game.spawner.RandomAsteroidSpeed()
	return spawning.AsteroidSpawnPlan{
		EntityType: spawning.SpawnEntityTypeAsteroid,
		Reason:     spawning.SpawnReasonDebugAsteroid,
		Position:   normalizedPosition,
		Velocity:   direction.Multiply(speed),
		Size:       rand.Intn(4) + 1,
		Variant:    rand.Intn(4),
	}
}
