package game

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/objectives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
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

func (game *Game) ForceLockFinalMatchState(endReason string) (FinalMatchState, bool) {
	game.mu.Lock()
	defer game.mu.Unlock()
	if game.lockedFinalMatchState != nil {
		return cloneFinalMatchState(*game.lockedFinalMatchState), true
	}
	if endReason == "" {
		endReason = "administratively_terminated"
	}
	if decision := game.evaluateMatchDecisionLocked(); decision.IsOver {
		return game.lockFinalMatchStateForDecisionLocked(decision)
	}

	facts := game.modeMatchFactsLocked()
	decision := rules.MatchDecision{
		IsOver:         true,
		TerminalStatus: rules.TerminalAdministrativelyTerminated,
		EndReason:      endReason,
		Players:        make([]rules.PlayerDecision, 0, len(facts.Players)),
	}
	for _, fact := range facts.Players {
		decision.Players = append(decision.Players, rules.PlayerDecision{
			ID: fact.ID, Status: fact.Status, Outcome: rules.OutcomeAborted,
		})
	}
	return game.lockFinalMatchStateForDecisionLocked(decision)
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
	playerSet := make(map[string]struct{}, len(game.participantRecords)+len(game.playerSessions))
	for playerID := range game.participantRecords {
		playerSet[playerID] = struct{}{}
	}
	for playerID := range game.playerSessions {
		playerSet[playerID] = struct{}{}
	}
	playerIDs := make([]string, 0, len(playerSet))
	for playerID := range playerSet {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Strings(playerIDs)

	facts := make([]PlayerMatchFact, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		teamID := teams.NoTeam
		score := 0
		if record := game.participantRecords[playerID]; record != nil {
			teamID = record.TeamID
			score = record.Score
		} else if session := game.playerSessions[playerID]; session != nil {
			teamID = session.TeamID
			score = session.Score
		}
		deathCount := 0
		if history, ok := game.lifeRuntime.DeathHistory(playerID); ok {
			deathCount = len(history)
		}
		facts = append(facts, PlayerMatchFact{
			GamePlayerID: playerID,
			TeamID:       teamID,
			Score:        score,
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
