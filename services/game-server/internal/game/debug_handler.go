package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
)

func (game *Game) handleDebugPacket(playerID string, player *entities.Ship, packet ClientPacket) bool {
	switch packet.Type {
	case PacketTypeDebugSpawnEntity:
		return game.handleDebugSpawnEntity(playerID, packet)
	case PacketTypeDebugRespawnPlayer:
		return game.handleDebugRespawnPlayer(playerID, packet)
	default:
		return false
	}
}
