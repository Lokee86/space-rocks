package modes

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
)

func EvaluateMatch(resolved ResolvedMatchRules, facts MatchFacts) rules.MatchDecision {
	if resolved.ModeID == ModeScoreAttack {
		return evaluateScoreAttack(resolved, facts)
	}
	return evaluateArcadeSurvival(facts)
}

func evaluateArcadeSurvival(facts MatchFacts) rules.MatchDecision {
	snapshot := rules.MatchSnapshot{HadParticipants: facts.HadParticipants}
	for _, fact := range facts.Players {
		if !fact.Active {
			continue
		}
		snapshot.Players = append(snapshot.Players, rules.PlayerSnapshot{
			ID:                fact.ID,
			HasActiveShip:     fact.Status == rules.PlayerActive,
			HasRemainingLives: fact.Status == rules.PlayerPendingRespawn,
		})
	}
	decision := rules.EvaluateMatch(snapshot)
	if decision.IsOver {
		decision.TerminalStatus = rules.TerminalCompleted
		decision.EndReason = string(EndNoActivePlayers)
	}
	return decision
}

func evaluateScoreAttack(resolved ResolvedMatchRules, facts MatchFacts) rules.MatchDecision {
	players := append([]PlayerFact(nil), facts.Players...)
	sort.Slice(players, func(left, right int) bool { return players[left].ID < players[right].ID })

	winnerIndex := -1
	for index, player := range players {
		if player.SuccessOrder <= 0 || player.Score < resolved.ObjectivePolicy.TargetScore {
			continue
		}
		if winnerIndex < 0 || player.SuccessOrder < players[winnerIndex].SuccessOrder {
			winnerIndex = index
		}
	}
	if winnerIndex >= 0 {
		decision := rules.MatchDecision{
			IsOver:           true,
			TerminalStatus:   rules.TerminalCompleted,
			EndReason:        string(EndTargetScoreReached),
			WinningPlayerIDs: []string{players[winnerIndex].ID},
		}
		for index, player := range players {
			outcome := rules.OutcomeLost
			placement := 0
			if index == winnerIndex {
				outcome = rules.OutcomeWon
				placement = 1
			}
			decision.Players = append(decision.Players, rules.PlayerDecision{
				ID: player.ID, Status: player.Status, Outcome: outcome, Placement: placement,
				CompletionTime: player.CompletionTime, TargetValue: float64(resolved.ObjectivePolicy.TargetScore),
			})
		}
		return decision
	}

	active := false
	decision := rules.MatchDecision{}
	for _, player := range players {
		if player.Active && player.Status != rules.PlayerEliminated {
			active = true
		}
		decision.Players = append(decision.Players, rules.PlayerDecision{
			ID: player.ID, Status: player.Status, TargetValue: float64(resolved.ObjectivePolicy.TargetScore),
		})
	}
	if facts.HadParticipants && !active {
		decision.IsOver = true
		decision.TerminalStatus = rules.TerminalFailed
		decision.EndReason = string(EndNoActivePlayers)
		for index := range decision.Players {
			decision.Players[index].Outcome = rules.OutcomeFailed
		}
	}
	return decision
}
