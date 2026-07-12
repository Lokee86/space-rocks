package game

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

	removed := len(target.game.entities.Asteroids)
	for id := range target.game.entities.Asteroids {
		delete(target.game.entities.Asteroids, id)
	}
	if removed > 0 {
		target.game.publishPresentationFrameLocked()
	}
	return removed
}
