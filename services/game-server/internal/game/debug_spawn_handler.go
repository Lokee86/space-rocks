package game

import (
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleDebugSpawnEntity(playerID string, packet ClientPacket) bool {
	request := devtools.SpawnEntityRequest{
		EntityType:     packet.EntityType,
		X:              packet.X,
		Y:              packet.Y,
		HasDirection:   packet.HasDirection,
		DirectionX:     packet.DirectionX,
		DirectionY:     packet.DirectionY,
		TargetPlayerID: packet.TargetPlayerID,
	}
	return game.DevtoolsHandleSpawnEntity(playerID, request)
}

func (game *Game) DevtoolsHandleSpawnEntity(playerID string, request devtools.SpawnEntityRequest) bool {
	if request.EntityType != devtools.EntityTypeAsteroid {
		logging.Game.Info("debug spawn entity not implemented for entity type",
			logging.FieldPlayerID, playerID,
			"entity_type", request.EntityType,
		)
		return true
	}
	return true
}
