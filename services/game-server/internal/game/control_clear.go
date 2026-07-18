package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/encounterlifecycle"

func (target *Control) ClearBullets() int {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()

	removed := len(target.game.entities.Projectiles)
	for id := range target.game.entities.Projectiles {
		delete(target.game.entities.Projectiles, id)
	}
	if removed > 0 {
		target.game.publishPresentationFrameLocked()
	}
	return removed
}

func (target *Control) ClearAsteroids() int {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()

	removed := target.game.removeAllAsteroidsForLifecycleTrigger(encounterlifecycle.TriggerScriptedCleanup)
	if removed > 0 {
		target.game.publishPresentationFrameLocked()
	}
	return removed
}
