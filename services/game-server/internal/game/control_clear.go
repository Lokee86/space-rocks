package game

func (target *Control) ClearBullets() int {
	removed := len(target.game.entities.Projectiles)
	for id := range target.game.entities.Projectiles {
		delete(target.game.entities.Projectiles, id)
	}
	return removed
}

func (target *Control) ClearAsteroids() int {
	removed := len(target.game.entities.Asteroids)
	for id := range target.game.entities.Asteroids {
		delete(target.game.entities.Asteroids, id)
	}
	return removed
}
