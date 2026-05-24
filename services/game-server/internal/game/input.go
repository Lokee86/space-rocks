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
	case PacketTypeToggleDebugInvincible:
		enabled := player.DevTools.ToggleInvincible()
		if session, ok := game.playerSessions[playerID]; ok {
			session.DevTools = player.DevTools
		}
		logging.Game.Info("debug invincibility toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
	case PacketTypeToggleDebugInfiniteLives:
		enabled := player.DevTools.ToggleInfiniteLives()
		if session, ok := game.playerSessions[playerID]; ok {
			session.DevTools = player.DevTools
		}
		logging.Game.Info("debug infinite lives toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
	case PacketTypeToggleDebugFreezeWorld:
		enabled := game.worldDevTools.ToggleFreezeWorld()
		logging.Game.Info("debug world freeze toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
	case PacketTypeToggleDebugFreezePlayer:
		enabled := player.DevTools.ToggleFreezePlayer()
		player.ClearInput()
		if session, ok := game.playerSessions[playerID]; ok {
			session.DevTools = player.DevTools
		}
		logging.Game.Info("debug player freeze toggled",
			logging.FieldPlayerID, playerID,
			"enabled", enabled,
		)
	case PacketTypeClientConfig:
		player.SetConfig(packet.Config)
	}
}
