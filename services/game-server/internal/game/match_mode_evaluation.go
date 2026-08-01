package game

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/modes"
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
	teamScores := make(map[teams.ID]int)
	for _, playerID := range playerIDs {
		status := rules.PlayerEliminated
		session, active := game.playerSessions[playerID]
		teamID := teams.NoTeam
		score := 0
		if session != nil {
			teamID = session.TeamID
			score = session.Score
			player, hasShip := game.entities.Players[playerID]
			if hasShip && player != nil && !player.IsPendingDespawn() {
				status = rules.PlayerActive
			} else if session.LifeOptions.InfiniteLives || session.Lives > 0 {
				status = rules.PlayerPendingRespawn
			}
		}
		if record := game.participantRecords[playerID]; record != nil {
			teamID = record.TeamID
			score = record.Score
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
		if teamID != teams.NoTeam {
			teamScores[teamID] += score
		}
	}
	if game.resolvedMatchRules.TeamScoreEnabled {
		teamIDs := make([]teams.ID, 0, len(teamScores))
		for teamID := range teamScores {
			teamIDs = append(teamIDs, teamID)
		}
		sort.Slice(teamIDs, func(left, right int) bool { return teamIDs[left] < teamIDs[right] })
		for _, teamID := range teamIDs {
			facts.Teams = append(facts.Teams, modes.TeamFact{
				ID:             teamID,
				Score:          teamScores[teamID],
				CompletionTime: game.teamScoreCompletionTimes[teamID],
				SuccessOrder:   game.teamScoreSuccessOrders[teamID],
			})
		}
	}
	return facts
}
