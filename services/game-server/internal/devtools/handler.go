package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func HandleCommand(target *game.Game, playerID string, command DebugCommand) bool {
	switch command.Type {
	case PacketTypeToggleDebugInvincible:
		return handleToggleDebugInvincible(target, playerID)
	case PacketTypeToggleDebugInfiniteLives:
		return handleToggleDebugInfiniteLives(target, playerID)
	case PacketTypeToggleDebugFreezeWorld:
		return handleToggleDebugFreezeWorld(target, playerID)
	case PacketTypeToggleDebugFreezePlayer:
		return handleToggleDebugFreezePlayer(target, playerID)
	case PacketTypeDebugKillPlayer:
		return handleDebugKillPlayer(target, playerID, command)
	case PacketTypeDebugSpawnEntity:
		return handleDebugSpawnEntity(target, playerID, command)
	case PacketTypeDebugRespawnPlayer:
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
		normalizedTargetPlayerID, spawnPosition, ok := applyDebugRespawnPlayer(target, request)
		if !ok {
			logging.Game.Info("debug respawn player ignored",
				logging.FieldPlayerID, playerID,
				"target_player_id", request.TargetPlayerID,
			)
			return true
		}
		logging.Game.Info("debug force respawn applied",
			logging.FieldPlayerID, playerID,
			"target_player_id", normalizedTargetPlayerID,
			"x", spawnPosition.X,
			"y", spawnPosition.Y,
		)
		return true
	default:
		return false
	}
}
