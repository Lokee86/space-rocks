package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func debugPickupSpawnPosition(request DebugCommand) physics.Vector2 {
	return space.NormalizePosition(physics.Vector2{X: request.X, Y: request.Y})
}

func handleDebugSpawnPickup(target *game.Game, playerID string, command DebugCommand) bool {
	pickupType := pickups.PickupType(command.PickupType)
	position := debugPickupSpawnPosition(command)
	pickup, ok, err := target.SpawnPickup(pickupType, position)
	if err != nil || !ok {
		logging.Game.Debug("debug pickup spawn ignored",
			logging.FieldPlayerID, playerID,
			"pickup_type", command.PickupType,
			"x", position.X,
			"y", position.Y,
		)
		return true
	}

	logging.Game.Debug("debug pickup spawned",
		logging.FieldPlayerID, playerID,
		"pickup_id", pickup.ID,
		"pickup_type", string(pickup.Type),
		"x", pickup.X,
		"y", pickup.Y,
	)
	return true
}
