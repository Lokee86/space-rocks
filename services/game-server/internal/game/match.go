package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"

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

func (game *Game) isMatchOverLocked() bool {
	return game.matchDecisionLocked().IsOver
}

func (game *Game) PlayerMatchFacts() []PlayerMatchFact {
	game.mu.Lock()
	defer game.mu.Unlock()

	facts := make([]PlayerMatchFact, 0, len(game.participantRecords))
	for _, record := range game.participantRecords {
		if record == nil {
			continue
		}
		facts = append(facts, PlayerMatchFact{
			GamePlayerID: record.ID,
			TeamID:       record.TeamID,
			Score:        record.Score,
			ShipDeaths:   record.ShipDeaths,
		})
	}
	return facts
}

func (game *Game) matchDecisionLocked() rules.MatchDecision {
	return rules.EvaluateMatch(game.matchSnapshot())
}

func (game *Game) matchSnapshot() rules.MatchSnapshot {
	players := make([]rules.PlayerSnapshot, 0, len(game.playerSessions))
	for playerID, session := range game.playerSessions {
		player, ok := game.entities.Players[playerID]
		hasActiveShip := ok && player != nil && !player.IsPendingDespawn()
		players = append(players, rules.PlayerSnapshot{
			ID:                session.ID,
			HasActiveShip:     hasActiveShip,
			HasRemainingLives: session.Lives > 0,
		})
	}
	return rules.MatchSnapshot{
		Players:         players,
		HadParticipants: len(game.participantRecords) > 0,
	}
}
