package game

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func (game *Game) evaluateMatchDecisionLocked() rules.MatchDecision {
	return modes.EvaluateMatch(game.resolvedMatchRules, game.modeMatchFactsLocked())
}

func (game *Game) modeMatchFactsLocked() modes.MatchFacts {
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

	facts := modes.MatchFacts{HadParticipants: len(playerIDs) > 0, Elapsed: game.matchElapsed}
	for _, playerID := range playerIDs {
		status := playerstate.StatusEliminated
		if lifecycle, ok := game.lifeRuntime.ParticipantSnapshot(playerID); ok {
			status = lifecycle.Status
		}
		session, active := game.playerSessions[playerID]
		teamID := teams.NoTeam
		score := 0
		if record := game.participantRecords[playerID]; record != nil {
			teamID = record.TeamID
			score = record.Score
		} else if session != nil {
			teamID = session.TeamID
			score = session.Score
		}
		facts.Players = append(facts.Players, modes.PlayerFact{
			ID:             playerID,
			TeamID:         teamID,
			Status:         status,
			Active:         active,
			Score:          score,
			CompletionTime: game.scoreCompletionTimes[playerID],
			SuccessOrder:   game.scoreSuccessOrders[playerID],
		})
	}
	return facts
}
