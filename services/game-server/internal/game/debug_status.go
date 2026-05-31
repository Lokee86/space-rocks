package game

func (game *Game) debugStatusFor(playerID string) DebugStatus {
	status := game.DevtoolsStatusFor(playerID)
	return DebugStatus{
		Invincible:    status.Invincible,
		InfiniteLives: status.InfiniteLives,
		WorldFrozen:   status.WorldFrozen,
		PlayerFrozen:  status.PlayerFrozen,
	}
}
