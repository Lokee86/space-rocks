package game

func (target *Control) SetPlayerScore(playerID string, score int) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	change := target.game.setPlayerScoreLocked(playerID, score)
	if change.Found && change.Delta != 0 {
		target.game.publishPresentationFrameLocked()
	}
	return change.Found
}

func (target *Control) AddPlayerScore(playerID string, amount int) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	change := target.game.addPlayerScoreLocked(playerID, amount)
	if change.Found && change.Delta != 0 {
		target.game.publishPresentationFrameLocked()
	}
	return change.Found
}

func (target *Control) SetPlayerLives(playerID string, lives int) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	change := target.game.setPlayerLivesLocked(playerID, lives)
	if change.Found && change.Delta != 0 {
		target.game.publishPresentationFrameLocked()
	}
	return change.Found
}

func (target *Control) AddPlayerLives(playerID string, amount int) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	change := target.game.addPlayerLivesLocked(playerID, amount)
	if change.Found && change.Delta != 0 {
		target.game.publishPresentationFrameLocked()
	}
	return change.Found
}
