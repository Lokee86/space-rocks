package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/motion"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func (game *Game) removeReadyPlayers() {
	for id, player := range game.state.Players {
		if player.ReadyForRemoval() {
			if session, ok := game.playerSessions[id]; ok {
				session.Score = player.Score
			}
			delete(game.state.Players, id)
		}
	}
}

func (game *Game) stepPlayerSessions(delta float64) {
	for _, session := range game.playerSessions {
		session.Step(delta)
	}
}

func (game *Game) stepPlayers(delta float64, bounds space.Bounds) {
	for _, player := range game.state.Players {
		motion.AdvanceShip(player, delta, bounds)
		if cameraView, ok := game.cameraViews[player.ID]; ok {
			cameraView.SetPosition(player.Position())
		}
		if player.IsPendingDespawn() {
			continue
		}
		if game.worldSimulationOptions.BulletsCanMove() && player.WantsToShoot() && player.CanShoot() {
			game.spawnBullet(player)
			player.ResetShootCooldown()
		}
	}
}
