package matchresults

import (
	"fmt"
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func ResolveDecision(input BuildInput) (MatchDecision, error) {
	if input.Session != SessionSinglePlayer && input.Session != SessionMultiplayer {
		return MatchDecision{}, fmt.Errorf("unsupported session context %q", input.Session)
	}
	participants := cloneParticipantFacts(input.Participants)
	sort.Slice(participants, func(left, right int) bool {
		if participants[left].Score != participants[right].Score {
			return participants[left].Score > participants[right].Score
		}
		return participants[left].PlayerRef.GamePlayerID < participants[right].PlayerRef.GamePlayerID
	})

	results := make([]ParticipantResult, len(participants))
	placements := scorePlacements(participants)
	for index, fact := range participants {
		disposition := fact.Disposition
		if disposition == "" {
			disposition = DispositionParticipated
		}
		results[index] = ParticipantResult{
			PlayerRef: fact.PlayerRef, TeamID: fact.TeamID, Placement: placements[index],
			Disposition: disposition, FinalScore: fact.Score, ShipDeaths: fact.ShipDeaths,
		}
	}

	decision := MatchDecision{TerminalStatus: TerminalCompleted, EndReason: input.EndReason, Participants: results}
	if decision.EndReason == "" {
		decision.EndReason = "simulation_complete"
	}
	resolveParticipantOutcomes(input.Session, &decision)
	decision.Teams, decision.WinningTeamRefs = resolveTeamOutcomes(input.TeamStructure, participants)
	return decision, nil
}

func resolveParticipantOutcomes(session SessionContext, decision *MatchDecision) {
	if session == SessionSinglePlayer {
		for index := range decision.Participants {
			decision.Participants[index].Outcome = OutcomeCompleted
		}
		return
	}
	if len(decision.Participants) == 0 {
		return
	}
	maxScore := decision.Participants[0].FinalScore
	winnerCount := 0
	for _, participant := range decision.Participants {
		if participant.FinalScore == maxScore {
			winnerCount++
		}
	}
	if winnerCount != 1 {
		for index := range decision.Participants {
			decision.Participants[index].Outcome = OutcomeDraw
		}
		return
	}
	for index := range decision.Participants {
		if decision.Participants[index].FinalScore == maxScore {
			decision.Participants[index].Outcome = OutcomeWon
			decision.WinningPlayerRefs = append(decision.WinningPlayerRefs, decision.Participants[index].PlayerRef)
		} else {
			decision.Participants[index].Outcome = OutcomeLost
		}
	}
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
	ids := make([]teams.ID, 0, len(totals))
	for id := range totals {
		ids = append(ids, id)
	}
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

func scorePlacements(participants []ParticipantFact) []int {
	placements := make([]int, len(participants))
	placement := 0
	lastScore := 0
	for index, participant := range participants {
		if index == 0 || participant.Score != lastScore {
			placement = index + 1
			lastScore = participant.Score
		}
		placements[index] = placement
	}
	return placements
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
