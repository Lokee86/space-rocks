package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"

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
	if _, ok := game.playerSessions[playerID]; !ok {
		return 0, false
	}
	value, _ := game.awardsRuntime().Counter(awards.Owner{Scope: awards.ScopePlayer, ID: playerID}, awards.CounterScore)
	return int(value), true
}

func (game *Game) setPlayerScoreLocked(playerID string, score int) PlayerCounterChange {
	before, found := game.currentPlayerScoreLocked(playerID)
	if !found {
		return PlayerCounterChange{PlayerID: playerID}
	}

	after := clampPlayerCounter(score)
	result, err := game.applyAwardMutationsLocked(game.nextAwardEventIDLocked("dev_score_set"), []awards.Mutation{{
		Owner: awards.Owner{Scope: awards.ScopePlayer, ID: playerID}, CounterID: awards.CounterScore,
		Operation: awards.MutationSet, Value: float64(after), Source: "devtools",
	}})
	if err != nil || !result.Applied || len(result.Changes) == 0 {
		return PlayerCounterChange{PlayerID: playerID, Found: true, Before: before, After: before}
	}
	change := result.Changes[0]
	return PlayerCounterChange{
		PlayerID: playerID,
		Found:    true,
		Before:   int(change.Before),
		After:    int(change.After),
		Delta:    int(change.Delta),
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
	if _, ok := game.lifeRuntime.ParticipantSnapshot(playerID); ok {
		lives, _ := game.lifeRuntime.ProjectedLives(playerID)
		return lives, true
	}

	return 0, false
}

func (game *Game) setPlayerLivesLocked(playerID string, lives int) PlayerCounterChange {
	mutation := game.lifeRuntime.SetLives(playerID, lives)
	if _, found := game.lifeRuntime.ParticipantSnapshot(playerID); !found {
		return PlayerCounterChange{PlayerID: playerID}
	}

	return PlayerCounterChange{
		PlayerID: playerID,
		Found:    true,
		Before:   mutation.PreviousLives,
		After:    mutation.CurrentLives,
		Delta:    mutation.Delta,
	}
}

func (game *Game) addPlayerLivesLocked(playerID string, amount int) PlayerCounterChange {
	mutation := game.lifeRuntime.AddLives(playerID, amount)
	if _, found := game.lifeRuntime.ParticipantSnapshot(playerID); !found {
		return PlayerCounterChange{PlayerID: playerID}
	}

	return PlayerCounterChange{
		PlayerID: playerID,
		Found:    true,
		Before:   mutation.PreviousLives,
		After:    mutation.CurrentLives,
		Delta:    mutation.Delta,
	}
}
