package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"

type simulationStepObserver func(float64)

func (target *Control) BulletsCanMove() bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.worldSimulationOptions.BulletsCanMove()
}

func (target *Control) SpawnDebugBullet(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	_, ok := target.game.spawnDebugBullet(ownerPlayerID, origin, direction)
	if ok {
		target.game.publishPresentationFrameLocked()
	}
	return ok
}

func (target *Control) RegisterSimulationStepObserverOnce(key string, observer func(float64, func() bool, func(string, physics.Vector2, physics.Vector2) bool)) {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	if key == "" || observer == nil {
		return
	}
	if _, registered := target.game.simulationStepObserverKeys[key]; registered {
		return
	}
	target.game.simulationStepObserverKeys[key] = struct{}{}

	target.game.simulationStepObservers = append(target.game.simulationStepObservers, func(delta float64) {
		observer(delta, func() bool {
			return target.game.worldSimulationOptions.BulletsCanMove()
		}, func(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
			_, ok := target.game.spawnDebugBullet(ownerPlayerID, origin, direction)
			return ok
		})
	})
}
