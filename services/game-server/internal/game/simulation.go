package game

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func (game *Game) runSimulation() {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	delta := 1.0 / float64(constants.ServerTickRate)
	for {
		select {
		case <-game.stopSimulation:
			return
		case <-ticker.C:
			game.Step(delta)
		}
	}
}

func (game *Game) Step(delta float64) {
	game.mu.Lock()
	defer game.mu.Unlock()

	bounds := space.DefaultBounds()

	game.stepPlayerSessions(delta)
	game.stepPlayers(delta, bounds)
	game.removeReadyPlayers()
	game.stepAsteroidSpawning(delta)
	game.stepAsteroids(delta, bounds)
	game.stepBullets(delta, bounds)
	game.stepCollisions()
}

func (game *Game) stepCollisions() {
	if game.worldDevTools.CanRunCollisions() {
		game.handleShipAsteroidCollisions()
		game.handleBulletAsteroidCollisions()
	}
}
