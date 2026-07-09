package game

func (target *Control) SetPlayerScore(playerID string, score int) bool {
	return target.game.SetPlayerScore(playerID, score).Found
}

func (target *Control) AddPlayerScore(playerID string, amount int) bool {
	return target.game.AddPlayerScore(playerID, amount).Found
}

func (target *Control) SetPlayerLives(playerID string, lives int) bool {
	return target.game.SetPlayerLives(playerID, lives).Found
}

func (target *Control) AddPlayerLives(playerID string, amount int) bool {
	return target.game.AddPlayerLives(playerID, amount).Found
}
