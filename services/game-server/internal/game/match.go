package game

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
)

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
		deathCount := 0
		if history, ok := game.lifeRuntime.DeathHistory(record.ID); ok {
			deathCount = len(history)
		}
		facts = append(facts, PlayerMatchFact{
			GamePlayerID: record.ID,
			TeamID:       record.TeamID,
			Score:        record.Score,
			ShipDeaths:   deathCount,
		})
	}
	return facts
}

func (game *Game) DeathHistory() []lives.DeathFact {
	game.mu.Lock()
	defer game.mu.Unlock()

	playerIDs := make([]string, 0, len(game.participantRecords))
	for playerID := range game.participantRecords {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)

	result := make([]lives.DeathFact, 0)
	for _, playerID := range playerIDs {
		history, ok := game.lifeRuntime.DeathHistory(playerID)
		if !ok {
			continue
		}
		result = append(result, history...)
	}
	return result
}

func (game *Game) MatchDeathFacts() []MatchDeathFact {
	return game.DeathHistory()
}

func (game *Game) matchDecisionLocked() rules.MatchDecision {
	return rules.EvaluateMatch(game.matchSnapshot())
}

func (game *Game) matchSnapshot() rules.MatchSnapshot {
	players := make([]rules.PlayerSnapshot, 0, len(game.playerSessions))
	for playerID, session := range game.playerSessions {
		lifecycle, ok := game.lifeRuntime.ParticipantSnapshot(playerID)
		if !ok {
			continue
		}
		player, ok := game.entities.Players[playerID]
		hasActiveShip := ok && player != nil && !player.IsPendingDespawn()
		players = append(players, rules.PlayerSnapshot{
			ID:                session.ID,
			Status:            lifecycle.Status,
			HasActiveShip:     hasActiveShip,
			HasRemainingLives: game.projectedPlayerLives(playerID, lifecycle) > 0,
		})
	}
	return rules.MatchSnapshot{
		Players:         players,
		HadParticipants: len(game.participantRecords) > 0,
	}
}
