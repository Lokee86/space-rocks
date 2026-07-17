package devtools

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

func debugPickupSpawnPosition(request DebugCommand) physics.Vector2 {
	return space.NormalizePosition(physics.Vector2{X: request.X, Y: request.Y})
}

func handleDebugSpawnPickup(target Target, playerID string, command DebugCommand) bool {
	pickupType := pickups.PickupType(command.PickupType)
	position := debugPickupSpawnPosition(command)
	_, ok, err := target.SpawnPickup(pickupType, position)
	if err != nil || !ok {
		return true
	}

	return true
}
