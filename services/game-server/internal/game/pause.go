package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) setPlayerPaused(playerID string, paused bool) {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return
	}
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
	session.Suspension.SetPaused(paused)
	if paused {
		player.ClearInput()
		player.Velocity = physics.Vector2{}
		logging.Game.Debug("player paused", logging.FieldPlayerID, playerID)
		return
	}
	player.ClearInput()
	player.InvulnerabilityRemaining = constants.PlayerResumeInvulnerabilitySeconds
	logging.Game.Debug("player resumed",
		logging.FieldPlayerID, playerID,
		"invulnerability", constants.PlayerResumeInvulnerabilitySeconds,
	)
}

func (game *Game) togglePlayerPaused(playerID string) {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return
	}
	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}
	if player.IsPendingDespawn() {
		return
	}
	game.setPlayerPaused(playerID, !session.Suspension.Paused)
}

func (game *Game) PlayerPauseStatePacket(playerID string) (PlayerPauseState, bool) {
	game.mu.Lock()
	defer game.mu.Unlock()

	session, ok := game.playerSessions[playerID]
	if !ok {
		return PlayerPauseState{}, false
	}
	if _, ok := game.state.Players[playerID]; !ok {
		return PlayerPauseState{}, false
	}
	return PlayerPauseState{
		Type:     PacketTypePlayerPauseState,
		PlayerID: playerID,
		Paused:   session.Suspension.Paused,
	}, true
}

func (game *Game) playerCanReceiveInput(playerID string, player *runtime.Ship) bool {
	if player.IsPendingDespawn() {
		return false
	}
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	return !session.Suspension.IsSuspended()
}

func (game *Game) playerCanMove(playerID string, player *runtime.Ship) bool {
	if player.IsPendingDespawn() {
		return false
	}
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	return !session.Suspension.IsSuspended()
}

func (game *Game) playerCanShoot(playerID string, player *runtime.Ship) bool {
	if player.IsPendingDespawn() {
		return false
	}
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	return !session.Suspension.IsSuspended() &&
		player.ShootCooldown == 0
}

func (game *Game) playerCanTakeCollisionDamage(playerID string, player *runtime.Ship) bool {
	if player.IsPendingDespawn() {
		return false
	}
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	return !session.Suspension.IsSuspended() &&
		!player.IsInvulnerable() &&
		player.DamageOptions.CanTakeDamage()
}

func (game *Game) playerCanReceiveScore(playerID string, player *runtime.Ship) bool {
	session, ok := game.playerSessions[playerID]
	if !ok {
		return false
	}
	return !session.Suspension.IsSuspended() &&
		!player.IsInvulnerable()
}
