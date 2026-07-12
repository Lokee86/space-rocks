package game

func (target *Control) SetPlayerScore(playerID string, score int) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.setPlayerScoreLocked(playerID, score).Found
}

func (target *Control) AddPlayerScore(playerID string, amount int) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.addPlayerScoreLocked(playerID, amount).Found
}

func (target *Control) SetPlayerLives(playerID string, lives int) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.setPlayerLivesLocked(playerID, lives).Found
}

func (target *Control) AddPlayerLives(playerID string, amount int) bool {
	target.game.mu.Lock()
	defer target.game.mu.Unlock()
	return target.game.addPlayerLivesLocked(playerID, amount).Found
}
