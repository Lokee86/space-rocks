package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/debugging"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleDebugSpawnEntity(playerID string, packet ClientPacket) bool {
	request := debugging.SpawnEntityRequest{
		EntityType:     packet.EntityType,
		X:              packet.X,
		Y:              packet.Y,
		HasDirection:   packet.HasDirection,
		DirectionX:     packet.DirectionX,
		DirectionY:     packet.DirectionY,
		TargetPlayerID: packet.TargetPlayerID,
	}
	if request.EntityType != debugging.EntityTypeAsteroid {
		logging.Game.Info("debug spawn entity not implemented for entity type",
			logging.FieldPlayerID, playerID,
			"entity_type", request.EntityType,
		)
		return true
	}

	plan := game.buildDebugAsteroidSpawnPlan(request)
	asteroid := game.applyAsteroidSpawn(plan)
	logging.Game.Info("debug asteroid spawned",
		logging.FieldPlayerID, playerID,
		"asteroid_id", asteroid.ID,
		"x", plan.Position.X,
		"y", plan.Position.Y,
		"has_direction", request.HasDirection,
	)
	return true
}
