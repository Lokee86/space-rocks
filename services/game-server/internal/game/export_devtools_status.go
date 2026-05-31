package game

type DevtoolsStatus struct {
	Invincible    bool
	InfiniteLives bool
	WorldFrozen   bool
	AsteroidsFrozen bool
	BulletsFrozen bool
	SpawningFrozen bool
	CollisionsFrozen bool
	PlayerFrozen  bool
}

func (game *Game) DevtoolsStatusFor(playerID string) DevtoolsStatus {
	status := DevtoolsStatus{
		WorldFrozen: game.worldSimulationOptions.IsWorldFrozen(),
		AsteroidsFrozen: !game.worldSimulationOptions.AsteroidsCanMove(),
		BulletsFrozen: !game.worldSimulationOptions.BulletsCanMove(),
		SpawningFrozen: !game.worldSimulationOptions.CanSpawnAsteroids(),
		CollisionsFrozen: !game.worldSimulationOptions.CanRunCollisions(),
	}

	if session, ok := game.playerSessions[playerID]; ok {
		status.Invincible = session.DamageOptions.Invincible
		status.InfiniteLives = session.LifeOptions.InfiniteLives
		status.PlayerFrozen = session.Suspension.DevFrozen
	}

	if player, ok := game.state.Players[playerID]; ok {
		status.Invincible = player.DamageOptions.Invincible
	}

	return status
}
