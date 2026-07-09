package game

import "github.com/Lokee86/space-rocks/server/internal/game/physics"

func (target *Control) BulletsCanMove() bool {
	return target.game.worldSimulationOptions.BulletsCanMove()
}

func (target *Control) SpawnDebugBullet(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	_, ok := target.SpawnBullet(ownerPlayerID, origin, direction)
	return ok
}

func (target *Control) RegisterSimulationStepObserver(observer func(float64)) {
	if observer == nil {
		return
	}

	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	target.game.simulationStepObservers = append(target.game.simulationStepObservers, observer)
}
