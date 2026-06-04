package game

func (game *Game) DevtoolsClearBullets() int {
	game.mu.Lock()
	defer game.mu.Unlock()

	removed := len(game.entities.Projectiles)
	for id := range game.entities.Projectiles {
		delete(game.entities.Projectiles, id)
	}

	return removed
}

func (game *Game) DevtoolsClearAsteroids() int {
	game.mu.Lock()
	defer game.mu.Unlock()

	removed := len(game.entities.Asteroids)
	for id := range game.entities.Asteroids {
		delete(game.entities.Asteroids, id)
	}

	return removed
}
