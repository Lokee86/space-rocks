package game

import "github.com/Lokee86/space-rocks/server/internal/game/rules"

func (game *Game) IsGameOver() bool {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.matchDecisionLocked().IsOver
}

func (game *Game) MatchDecision() rules.MatchDecision {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.matchDecisionLocked()
}

func (game *Game) matchDecisionLocked() rules.MatchDecision {
	return rules.EvaluateMatch(game.matchSnapshot())
}

func (game *Game) matchSnapshot() rules.MatchSnapshot {
	players := make([]rules.PlayerSnapshot, 0, len(game.playerSessions))
	for playerID, session := range game.playerSessions {
		player, ok := game.state.Players[playerID]
		hasActiveShip := ok && player != nil && !player.IsPendingDespawn()
		players = append(players, rules.PlayerSnapshot{
			ID:                session.ID,
			HasActiveShip:     hasActiveShip,
			HasRemainingLives: session.Lives > 0,
		})
	}
	return rules.MatchSnapshot{Players: players}
}
