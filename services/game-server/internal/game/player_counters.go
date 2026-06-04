package game

type PlayerCounterChange struct {
	PlayerID string
	Found    bool
	Before   int
	After    int
	Delta    int
}

func (game *Game) SetPlayerScore(playerID string, score int) PlayerCounterChange {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.setPlayerScoreLocked(playerID, score)
}

func (game *Game) AddPlayerScore(playerID string, amount int) PlayerCounterChange {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.addPlayerScoreLocked(playerID, amount)
}

func (game *Game) SetPlayerLives(playerID string, lives int) PlayerCounterChange {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.setPlayerLivesLocked(playerID, lives)
}

func (game *Game) AddPlayerLives(playerID string, amount int) PlayerCounterChange {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.addPlayerLivesLocked(playerID, amount)
}

func clampPlayerCounter(value int) int {
	if value < 0 {
		return 0
	}

	return value
}

func (game *Game) currentPlayerScoreLocked(playerID string) (int, bool) {
	if player, ok := game.entities.Players[playerID]; ok {
		return player.Score, true
	}
	if session, ok := game.playerSessions[playerID]; ok {
		return session.Score, true
	}

	return 0, false
}

func (game *Game) setPlayerScoreLocked(playerID string, score int) PlayerCounterChange {
	before, found := game.currentPlayerScoreLocked(playerID)
	if !found {
		return PlayerCounterChange{PlayerID: playerID}
	}

	after := clampPlayerCounter(score)
	if session, ok := game.playerSessions[playerID]; ok {
		session.Score = after
	}
	if player, ok := game.entities.Players[playerID]; ok {
		player.Score = after
	}

	return PlayerCounterChange{
		PlayerID: playerID,
		Found:    true,
		Before:   before,
		After:    after,
		Delta:    after - before,
	}
}

func (game *Game) addPlayerScoreLocked(playerID string, amount int) PlayerCounterChange {
	before, found := game.currentPlayerScoreLocked(playerID)
	if !found {
		return PlayerCounterChange{PlayerID: playerID}
	}

	return game.setPlayerScoreLocked(playerID, before+amount)
}

func (game *Game) currentPlayerLivesLocked(playerID string) (int, bool) {
	if session, ok := game.playerSessions[playerID]; ok {
		return session.Lives, true
	}
	if player, ok := game.entities.Players[playerID]; ok {
		return player.Lives, true
	}

	return 0, false
}

func (game *Game) setPlayerLivesLocked(playerID string, lives int) PlayerCounterChange {
	before, found := game.currentPlayerLivesLocked(playerID)
	if !found {
		return PlayerCounterChange{PlayerID: playerID}
	}

	after := clampPlayerCounter(lives)
	if session, ok := game.playerSessions[playerID]; ok {
		session.Lives = after
	}
	if player, ok := game.entities.Players[playerID]; ok {
		player.Lives = after
	}

	return PlayerCounterChange{
		PlayerID: playerID,
		Found:    true,
		Before:   before,
		After:    after,
		Delta:    after - before,
	}
}

func (game *Game) addPlayerLivesLocked(playerID string, amount int) PlayerCounterChange {
	before, found := game.currentPlayerLivesLocked(playerID)
	if !found {
		return PlayerCounterChange{PlayerID: playerID}
	}

	return game.setPlayerLivesLocked(playerID, before+amount)
}
