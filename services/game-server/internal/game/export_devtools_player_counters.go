package game

func (game *Game) DevtoolsSetPlayerScore(playerID string, score int) PlayerCounterChange {
	return game.SetPlayerScore(playerID, score)
}

func (game *Game) DevtoolsAddPlayerScore(playerID string, amount int) PlayerCounterChange {
	return game.AddPlayerScore(playerID, amount)
}

func (game *Game) DevtoolsSetPlayerLives(playerID string, lives int) PlayerCounterChange {
	return game.SetPlayerLives(playerID, lives)
}

func (game *Game) DevtoolsAddPlayerLives(playerID string, amount int) PlayerCounterChange {
	return game.AddPlayerLives(playerID, amount)
}
