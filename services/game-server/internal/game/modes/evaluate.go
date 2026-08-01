package modes

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
)

func EvaluateMatch(resolved ResolvedMatchRules, facts MatchFacts) rules.MatchDecision {
	switch resolved.ModeID {
	case ModeScoreAttack:
		return evaluateScoreAttack(resolved, facts)
	case ModeDeathmatch:
		return evaluateDeathmatch(resolved, facts)
	default:
		return evaluateArcadeSurvival(facts)
	}
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

func evaluateDeathmatch(resolved ResolvedMatchRules, facts MatchFacts) rules.MatchDecision {
	if resolved.TeamScoreEnabled {
		return evaluateTeamDeathmatch(resolved, facts)
	}

	players := append([]PlayerFact(nil), facts.Players...)
	sort.Slice(players, func(left, right int) bool { return players[left].ID < players[right].ID })

	winnerIndex := -1
	for index, player := range players {
		if player.SuccessOrder <= 0 || player.Score < resolved.ObjectivePolicy.TargetKills {
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
			EndReason:        string(EndTargetKillsReached),
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
				CompletionTime: player.CompletionTime, TargetValue: float64(resolved.ObjectivePolicy.TargetKills),
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
			ID: player.ID, Status: player.Status, TargetValue: float64(resolved.ObjectivePolicy.TargetKills),
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

func evaluateTeamDeathmatch(resolved ResolvedMatchRules, facts MatchFacts) rules.MatchDecision {
	players := append([]PlayerFact(nil), facts.Players...)
	sort.Slice(players, func(left, right int) bool { return players[left].ID < players[right].ID })
	teamFacts := append([]TeamFact(nil), facts.Teams...)
	sort.Slice(teamFacts, func(left, right int) bool { return teamFacts[left].ID < teamFacts[right].ID })

	winnerIndex := -1
	for index, team := range teamFacts {
		if team.SuccessOrder <= 0 || team.Score < resolved.ObjectivePolicy.TargetKills {
			continue
		}
		if winnerIndex < 0 || team.SuccessOrder < teamFacts[winnerIndex].SuccessOrder {
			winnerIndex = index
		}
	}
	if winnerIndex >= 0 {
		winnerID := teamFacts[winnerIndex].ID
		decision := rules.MatchDecision{
			IsOver:         true,
			TerminalStatus: rules.TerminalCompleted,
			EndReason:      string(EndTargetKillsReached),
		}
		for _, player := range players {
			outcome := rules.OutcomeLost
			placement := 0
			completionTime := 0.0
			if player.TeamID == winnerID {
				outcome = rules.OutcomeWon
				placement = 1
				completionTime = teamFacts[winnerIndex].CompletionTime
				decision.WinningPlayerIDs = append(decision.WinningPlayerIDs, player.ID)
			}
			decision.Players = append(decision.Players, rules.PlayerDecision{
				ID: player.ID, Status: player.Status, Outcome: outcome, Placement: placement,
				CompletionTime: completionTime, TargetValue: float64(resolved.ObjectivePolicy.TargetKills),
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
			ID: player.ID, Status: player.Status, TargetValue: float64(resolved.ObjectivePolicy.TargetKills),
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
