package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (game *Game) DevtoolsBulletsCanMove() bool {
	return game.worldSimulationOptions.BulletsCanMove()
}

func (game *Game) DevtoolsSpawnDebugBullet(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	_, spawned := game.spawnDebugBullet(ownerPlayerID, origin, direction)
	return spawned
}

func (game *Game) DevtoolsRegisterSimulationStepObserver(observer func(float64)) {
	if observer == nil {
		return
	}
	game.simulationStepObservers = append(game.simulationStepObservers, observer)
}
