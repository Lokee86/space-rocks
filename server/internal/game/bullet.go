package game

func (bullet *Bullet) State() BulletState {
	return BulletState{
		ID:       bullet.ID,
		OwnerID:  bullet.OwnerID,
		X:        bullet.X,
		Y:        bullet.Y,
		Rotation: bullet.Rotation,
	}
}

func (bullet *Bullet) step(delta float64) {
	if bullet.PendingDespawn {
		bullet.DespawnDelay -= delta
		return
	}

	bullet.X += bullet.Velocity.X * delta
	bullet.Y += bullet.Velocity.Y * delta
	bullet.Life -= delta
}

func (bullet *Bullet) collisionBody(catalog CollisionShapeCatalog) (CollisionBody, bool) {
	shape, err := catalog.BulletShape()
	if err != nil {
		return CollisionBody{}, false
	}

	return CollisionBody{
		ID:       bullet.ID,
		Position: Vector2{X: bullet.X, Y: bullet.Y},
		Rotation: bullet.Rotation,
		Shape:    shape,
	}, true
}
