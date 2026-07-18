package matchresults

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func resolveLockedTeamOutcomes(structure teams.Structure, participants []ParticipantResult) ([]TeamResult, []teams.ID) {
	if structure == teams.StructureFFA || len(participants) == 0 {
		return nil, nil
	}
	totals := make(map[teams.ID]int)
	outcomes := make(map[teams.ID]Outcome)
	for _, participant := range participants {
		if participant.TeamID == teams.NoTeam {
			continue
		}
		totals[participant.TeamID] += participant.FinalScore
		outcomes[participant.TeamID] = mergeTeamOutcome(outcomes[participant.TeamID], participant.Outcome)
	}
	ids := sortedTeamIDs(totals)
	results := make([]TeamResult, 0, len(ids))
	winners := make([]teams.ID, 0)
	for _, id := range ids {
		outcome := outcomes[id]
		placement := 0
		if outcome == OutcomeWon || (structure == teams.StructureCoOp && outcome == OutcomeCompleted) {
			placement = 1
			winners = append(winners, id)
		}
		results = append(results, TeamResult{TeamID: id, Outcome: outcome, Placement: placement, FinalScore: totals[id]})
	}
	return results, winners
}

func mergeTeamOutcome(current Outcome, next Outcome) Outcome {
	priority := map[Outcome]int{
		OutcomeWon: 6, OutcomeCompleted: 5, OutcomeDraw: 4,
		OutcomeLost: 3, OutcomeFailed: 2, OutcomeAborted: 1,
	}
	if priority[next] > priority[current] {
		return next
	}
	return current
}

func resolveTeamOutcomes(structure teams.Structure, participants []ParticipantFact) ([]TeamResult, []teams.ID) {
	if structure == teams.StructureFFA || len(participants) == 0 {
		return nil, nil
	}
	totals := make(map[teams.ID]int)
	for _, participant := range participants {
		if participant.TeamID != teams.NoTeam {
			totals[participant.TeamID] += participant.Score
		}
	}
	if len(totals) == 0 {
		return nil, nil
	}
	ids := sortedTeamIDs(totals)
	sort.Slice(ids, func(left, right int) bool {
		if totals[ids[left]] != totals[ids[right]] {
			return totals[ids[left]] > totals[ids[right]]
		}
		return ids[left] < ids[right]
	})
	results := make([]TeamResult, len(ids))
	if structure == teams.StructureCoOp {
		for index, id := range ids {
			results[index] = TeamResult{TeamID: id, Outcome: OutcomeCompleted, Placement: 1, FinalScore: totals[id]}
		}
		return results, append([]teams.ID(nil), ids...)
	}
	maxScore := totals[ids[0]]
	maxCount := 0
	for _, id := range ids {
		if totals[id] == maxScore {
			maxCount++
		}
	}
	placements := teamPlacements(ids, totals)
	winners := make([]teams.ID, 0, 1)
	for index, id := range ids {
		outcome := OutcomeLost
		if maxCount != 1 {
			outcome = OutcomeDraw
		} else if totals[id] == maxScore {
			outcome = OutcomeWon
			winners = append(winners, id)
		}
		results[index] = TeamResult{TeamID: id, Outcome: outcome, Placement: placements[index], FinalScore: totals[id]}
	}
	return results, winners
}

func sortedTeamIDs(totals map[teams.ID]int) []teams.ID {
	ids := make([]teams.ID, 0, len(totals))
	for id := range totals {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left, right int) bool { return ids[left] < ids[right] })
	return ids
}

func teamPlacements(ids []teams.ID, totals map[teams.ID]int) []int {
	placements := make([]int, len(ids))
	placement := 0
	lastScore := 0
	for index, id := range ids {
		if index == 0 || totals[id] != lastScore {
			placement = index + 1
			lastScore = totals[id]
		}
		placements[index] = placement
	}
	return placements
}
