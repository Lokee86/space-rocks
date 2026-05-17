package game

func (asteroid *Asteroid) State() AsteroidState {
	return AsteroidState{
		ID:      asteroid.ID,
		X:       asteroid.X,
		Y:       asteroid.Y,
		Size:    asteroid.Size,
		Variant: asteroid.Variant,
	}
}

func (asteroid *Asteroid) step(delta float64) {
	asteroid.X += asteroid.Velocity.X * delta
	asteroid.Y += asteroid.Velocity.Y * delta
}
