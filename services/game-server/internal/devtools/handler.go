package devtools

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
)

func HandleCommand(target *game.Game, playerID string, command DebugCommand) bool {
	switch command.Type {
	case PacketTypeToggleDebugInvincible:
		return handleToggleDebugInvincible(target, playerID, command)
	case PacketTypeToggleDebugInfiniteLives:
		return handleToggleDebugInfiniteLives(target, playerID, command)
	case PacketTypeToggleDebugFreezeWorld:
		return handleToggleDebugFreezeWorld(target, playerID, command)
	case PacketTypeToggleDebugFreezePlayer:
		return handleToggleDebugFreezePlayer(target, playerID, command)
	case PacketTypeDebugKillPlayer:
		return handleDebugKillPlayer(target, playerID, command)
	case PacketTypeDebugSpawnEntity:
		return handleDebugSpawnEntity(target, playerID, command)
	case PacketTypeDebugRespawnPlayer:
		return handleDebugRespawnPlayer(target, playerID, command)
	case PacketTypeDebugSetScore:
		return handleDebugSetScore(target, playerID, command)
	case PacketTypeDebugAddScore:
		return handleDebugAddScore(target, playerID, command)
	case PacketTypeDebugSetLives:
		return handleDebugSetLives(target, playerID, command)
	case PacketTypeDebugAddLives:
		return handleDebugAddLives(target, playerID, command)
	case PacketTypeDebugClearBullets:
		return handleDebugClearBullets(target, playerID, command)
	case PacketTypeDebugClearAsteroids:
		return handleDebugClearAsteroids(target, playerID, command)
	default:
		return false
	}
}
