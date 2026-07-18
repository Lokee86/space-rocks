package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/motion"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
)

func (game *Game) removeReadyPlayers() {
	for id, player := range game.entities.Players {
		if player.ReadyForRemoval() {
			delete(game.entities.Players, id)
		}
	}
}

func (game *Game) stepPlayerSessions(delta float64) {
	game.lifeRuntime.Step(delta)
	for _, request := range game.participationRuntime.Step(delta, game.lifeRuntime.Status) {
		game.lifeRuntime.RemoveParticipant(request.PlayerID, request.ReasonCode)
		game.participationRuntime.UnregisterParticipant(request.PlayerID)
		game.removeActivePlayerLocked(request.PlayerID)
	}
}

func (game *Game) stepPlayers(delta float64, bounds space.Bounds) {
	for _, player := range game.entities.Players {
		motion.AdvanceShipWithMovePolicy(player, delta, bounds, game.playerCanMove(player.ID, player))
		if cameraView, ok := game.cameraViews[player.ID]; ok {
			cameraView.SetPosition(player.Position())
		}
		if player.IsPendingDespawn() {
			continue
		}
		if game.worldSimulationOptions.BulletsCanMove() && player.Input.PrimaryFire && game.playerCanShoot(player.ID, player) {
			game.firePlayerPrimaryWeapon(player.ID, player)
		}
		if game.worldSimulationOptions.BulletsCanMove() && player.Input.SecondaryFire && game.playerCanShoot(player.ID, player) {
			game.firePlayerSecondaryWeapon(player.ID, player)
		}
	}
}
