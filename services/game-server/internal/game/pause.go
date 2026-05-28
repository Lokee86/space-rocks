package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) setPlayerPaused(playerID string, paused bool) {
	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}
	if player.IsPendingDespawn() {
		if !paused {
			logging.Game.Debug("resume ignored; player pending despawn", logging.FieldPlayerID, playerID)
		}
		return
	}
	if paused {
		player.Pause()
		logging.Game.Debug("player paused", logging.FieldPlayerID, playerID)
		return
	}
	player.Resume(constants.PlayerResumeInvulnerabilitySeconds)
	logging.Game.Debug("player resumed",
		logging.FieldPlayerID, playerID,
		"invulnerability", constants.PlayerResumeInvulnerabilitySeconds,
	)
}

func (game *Game) togglePlayerPaused(playerID string) {
	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}
	if player.IsPendingDespawn() {
		return
	}
	game.setPlayerPaused(playerID, !player.Suspension.Paused)
}
