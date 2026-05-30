package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/debugging"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleDebugRespawnPlayer(playerID string, packet ClientPacket) bool {
	request := debugging.RespawnPlayerRequest{
		TargetPlayerID: packet.TargetPlayerID,
		X:              packet.X,
		Y:              packet.Y,
	}

	logging.Game.Info("debug respawn player received",
		logging.FieldPlayerID, playerID,
		"target_player_id", request.TargetPlayerID,
		"x", request.X,
		"y", request.Y,
	)
	return true
}
