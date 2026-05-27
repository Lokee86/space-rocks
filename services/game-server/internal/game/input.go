package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) HandlePacket(playerID string, packet ClientPacket) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if packet.Type == PacketTypeRespawn {
		game.respawnPlayer(playerID)
		return
	}
	if packet.Type == PacketTypeClientConfig {
		if session, ok := game.playerSessions[playerID]; ok {
			session.Config = packet.Config
		}
		if cameraView, ok := game.cameraViews[playerID]; ok {
			cameraView.SetConfig(packet.Config)
		}
	}

	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}
	if game.handleDebugPacket(playerID, player, packet) {
		return
	}
	switch packet.Type {
	case PacketTypeInput:
		if !player.CanReceiveInput() {
			return
		}
		player.SetInput(packet.Input)
	case PacketTypePausePlayer:
		if player.IsPendingDespawn() {
			return
		}
		player.Pause()
		logging.Game.Debug("player paused", logging.FieldPlayerID, playerID)
	case PacketTypeResumePlayer:
		if player.IsPendingDespawn() {
			logging.Game.Debug("resume ignored; player pending despawn", logging.FieldPlayerID, playerID)
			return
		}
		player.Resume(constants.PlayerResumeInvulnerabilitySeconds)
		logging.Game.Debug("player resumed",
			logging.FieldPlayerID, playerID,
			"invulnerability", constants.PlayerResumeInvulnerabilitySeconds,
		)
	case PacketTypeClientConfig:
		player.SetConfig(packet.Config)
	}
}
