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
	if asteroid.PendingDespawn {
		asteroid.DespawnDelay -= delta
		return
	}

	asteroid.X += asteroid.Velocity.X * delta
	asteroid.Y += asteroid.Velocity.Y * delta
}

func (asteroid *Asteroid) collisionBody(catalog CollisionShapeCatalog) (CollisionBody, bool) {
	shape, err := catalog.AsteroidShape(asteroid.Variant, asteroid.Size)
	if err != nil {
		return CollisionBody{}, false
	}

	return CollisionBody{
		ID:       asteroid.ID,
		Position: Vector2{X: asteroid.X, Y: asteroid.Y},
		Shape:    shape,
	}, true
}
