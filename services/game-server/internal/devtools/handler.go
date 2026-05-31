package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
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
		return handleDebugRespawnPlayer(target, playerID, command)
	default:
		return false
	}
}
