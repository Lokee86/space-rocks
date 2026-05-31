package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func handleDebugRespawnPlayer(target *game.Game, playerID string, command DebugCommand) bool {
	request := RespawnPlayerRequest{
		TargetPlayerID: command.TargetPlayerID,
		X:              command.X,
		Y:              command.Y,
	}
	logging.Game.Info("debug respawn player received",
		logging.FieldPlayerID, playerID,
		"target_player_id", request.TargetPlayerID,
		"x", request.X,
		"y", request.Y,
	)

	if target == nil {
		logging.Game.Info("debug respawn player ignored",
			logging.FieldPlayerID, playerID,
			"target_player_id", request.TargetPlayerID,
		)
		return true
	}

	normalizedTargetPlayerID, ok := resolveDebugRespawnTargetPlayerID(request)
	if !ok {
		logging.Game.Info("debug respawn player ignored",
			logging.FieldPlayerID, playerID,
			"target_player_id", request.TargetPlayerID,
		)
		return true
	}

	isPlayerAlive := false
	for _, player := range target.MatchDecision().Players {
		if player.ID == normalizedTargetPlayerID {
			isPlayerAlive = player.Status == "active"
			break
		}
	}

	if isPlayerAlive {
		logging.Game.Info("debug respawn player ignored",
			logging.FieldPlayerID, playerID,
			"target_player_id", normalizedTargetPlayerID,
		)
		return true
	}

	request.TargetPlayerID = normalizedTargetPlayerID

	respawnedPlayerID, spawnPosition, ok := applyDebugRespawnPlayer(target, request)
	if !ok {
		logging.Game.Info("debug respawn player ignored",
			logging.FieldPlayerID, playerID,
			"target_player_id", request.TargetPlayerID,
		)
		return true
	}

	logging.Game.Info("debug force respawn applied",
		logging.FieldPlayerID, playerID,
		"target_player_id", respawnedPlayerID,
		"x", spawnPosition.X,
		"y", spawnPosition.Y,
	)
	return true
}