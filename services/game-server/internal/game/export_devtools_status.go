package game

type DevtoolsStatus struct {
	Invincible    bool
	InfiniteLives bool
	WorldFrozen   bool
	PlayerFrozen  bool
}

func (game *Game) DevtoolsStatusFor(playerID string) DevtoolsStatus {
	status := DevtoolsStatus{
		WorldFrozen: game.worldSimulationOptions.IsWorldFrozen(),
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
