package game

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/objectives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
)

func (game *Game) IsGameOver() bool {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.isMatchOverLocked()
}

func (game *Game) MatchDecision() rules.MatchDecision {
	game.mu.Lock()
	defer game.mu.Unlock()
	return cloneMatchDecision(game.matchDecisionLocked())
}

func (game *Game) LockFinalMatchState() (FinalMatchState, bool) {
	game.mu.Lock()
	defer game.mu.Unlock()
	return game.lockFinalMatchStateLocked()
}

func (game *Game) LockedFinalMatchState() (FinalMatchState, bool) {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState == nil {
		return FinalMatchState{}, false
	}
	return cloneFinalMatchState(*game.lockedFinalMatchState), true
}

func (game *Game) isMatchOverLocked() bool {
	if game.lockedFinalMatchState != nil {
		return true
	}
	decision := game.evaluateMatchDecisionLocked()
	if !decision.IsOver {
		return false
	}
	_, _ = game.lockFinalMatchStateForDecisionLocked(decision)
	return true
}

func (game *Game) PlayerMatchFacts() []PlayerMatchFact {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState != nil {
		return append([]PlayerMatchFact(nil), game.lockedFinalMatchState.Players...)
	}
	return game.playerMatchFactsLocked()
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
	if game.lockedFinalMatchState != nil {
		return game.lockedFinalMatchState.Decision
	}
	return game.evaluateMatchDecisionLocked()
}

func (game *Game) lockFinalMatchStateLocked() (FinalMatchState, bool) {
	if game.lockedFinalMatchState != nil {
		return cloneFinalMatchState(*game.lockedFinalMatchState), true
	}
	decision := game.evaluateMatchDecisionLocked()
	if !decision.IsOver {
		return FinalMatchState{}, false
	}
	return game.lockFinalMatchStateForDecisionLocked(decision)
}

func (game *Game) lockFinalMatchStateForDecisionLocked(decision rules.MatchDecision) (FinalMatchState, bool) {
	if game.lockedFinalMatchState != nil {
		return cloneFinalMatchState(*game.lockedFinalMatchState), true
	}
	state := FinalMatchState{
		MatchID:       game.matchID,
		TraceID:       game.matchTraceID,
		ModeID:        game.modeID,
		TeamStructure: game.teamStructure,
		ResolvedRules: modes.CloneResolvedMatchRules(game.resolvedMatchRules),
		Decision:      cloneMatchDecision(decision),
		Players:       game.playerMatchFactsLocked(),
		Awards:        game.gameplayAwardSnapshotLocked(),
		Objectives:    game.objectivesRuntime().Snapshots(objectives.Viewer{IncludeHidden: true}),
	}
	game.lockedFinalMatchState = &state
	return cloneFinalMatchState(state), true
}

func (game *Game) playerMatchFactsLocked() []PlayerMatchFact {
	playerIDs := make([]string, 0, len(game.participantRecords))
	for playerID := range game.participantRecords {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)

	facts := make([]PlayerMatchFact, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		record := game.participantRecords[playerID]
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

func cloneMatchDecision(source rules.MatchDecision) rules.MatchDecision {
	clone := source
	clone.Players = append([]rules.PlayerDecision(nil), source.Players...)
	clone.WinningPlayerIDs = append([]string(nil), source.WinningPlayerIDs...)
	return clone
}
