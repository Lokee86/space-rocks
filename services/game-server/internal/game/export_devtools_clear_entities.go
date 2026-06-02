package game

func (game *Game) DevtoolsClearBullets() int {
	game.mu.Lock()
	defer game.mu.Unlock()

	removed := len(game.state.Projectiles)
	for id := range game.state.Projectiles {
		delete(game.state.Projectiles, id)
	}
	game.DevtoolsClearContinuousBulletStreams()

	return removed
}

func (game *Game) DevtoolsClearAsteroids() int {
	game.mu.Lock()
	defer game.mu.Unlock()

	removed := len(game.state.Asteroids)
	for id := range game.state.Asteroids {
		delete(game.state.Asteroids, id)
	}

	return removed
}
