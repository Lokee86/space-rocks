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
	bullet.X += bullet.Velocity.X * delta
	bullet.Y += bullet.Velocity.Y * delta
	bullet.Life -= delta
}
