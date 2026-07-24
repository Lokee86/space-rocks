package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/bots"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func (game *Game) AddBot() string {
	game.mu.Lock()
	defer game.mu.Unlock()

	playerID := game.addPlayerLocked()
	game.enableBotPlayerLocked(playerID)
	return playerID
}

func (game *Game) enableBotPlayerLocked(playerID string) bool {
	if playerID == "" || game.playerSessions[playerID] == nil {
		return false
	}
	game.botControllers[playerID] = bots.NewController()
	return true
}

func (game *Game) stepBots() {
	for playerID, controller := range game.botControllers {
		ship, ok := game.entities.Players[playerID]
		if !ok || ship == nil {
			session := game.playerSessions[playerID]
			if session == nil || !session.CanRespawn() {
				continue
			}
			game.respawnPlayer(playerID)
			ship = game.entities.Players[playerID]
		}
		if ship == nil || !game.playerCanReceiveInput(playerID, ship) {
			continue
		}

		observation := bots.Observation{
			Position:  ship.Position(),
			Velocity:  ship.Velocity,
			Rotation:  ship.Rotation,
			Asteroids: make([]bots.AsteroidObservation, 0, len(game.entities.Asteroids)),
		}
		for _, asteroid := range game.entities.Asteroids {
			if asteroid == nil || asteroid.PendingDespawn {
				continue
			}
			observation.Asteroids = append(observation.Asteroids, bots.AsteroidObservation{
				Position: physics.Vector2{X: asteroid.X, Y: asteroid.Y},
				Velocity: asteroid.Velocity,
				Size:     asteroid.Size,
			})
		}
		ship.SetInput(controller.Decide(observation))
	}
}
