package game

import (
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) handleDebugRespawnPlayer(playerID string, packet ClientPacket) bool {
	request := devtools.RespawnPlayerRequest{
		TargetPlayerID: packet.TargetPlayerID,
		X:              packet.X,
		Y:              packet.Y,
	}
	return game.DevtoolsHandleRespawnPlayer(playerID, request)
}

func (game *Game) DevtoolsHandleRespawnPlayer(playerID string, request devtools.RespawnPlayerRequest) bool {
	logging.Game.Info("debug respawn player received",
		logging.FieldPlayerID, playerID,
		"target_player_id", request.TargetPlayerID,
		"x", request.X,
		"y", request.Y,
	)

	if !game.applyDebugRespawnPlayer(request) {
		logging.Game.Info("debug respawn player ignored",
			logging.FieldPlayerID, playerID,
			"target_player_id", request.TargetPlayerID,
		)
	}

	return true
}
