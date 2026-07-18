package matchresults

import (
	"fmt"
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"
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
	if input.LockedDecision.TerminalStatus != "" {
		decision.TerminalStatus = TerminalStatus(input.LockedDecision.TerminalStatus)
	}
	if input.LockedDecision.EndReason != "" {
		decision.EndReason = input.LockedDecision.EndReason
	}
	if decision.EndReason == "" {
		decision.EndReason = "simulation_complete"
	}

	lockedResults := applyLockedParticipantDecision(&decision, input.LockedDecision)
	if !lockedResults {
		resolveParticipantOutcomes(input.Session, &decision)
	}
	if lockedResults {
		decision.Teams, decision.WinningTeamRefs = resolveLockedTeamOutcomes(input.TeamStructure, decision.Participants)
	} else {
		decision.Teams, decision.WinningTeamRefs = resolveTeamOutcomes(input.TeamStructure, participants)
	}
	return decision, nil
}

func applyLockedParticipantDecision(decision *MatchDecision, locked rules.MatchDecision) bool {
	byID := make(map[string]int, len(decision.Participants))
	for index, participant := range decision.Participants {
		byID[participant.PlayerRef.GamePlayerID] = index
	}
	applied := false
	for _, player := range locked.Players {
		index, ok := byID[player.ID]
		if !ok {
			continue
		}
		if player.Outcome != "" {
			decision.Participants[index].Outcome = Outcome(player.Outcome)
			applied = true
		}
		decision.Participants[index].Placement = player.Placement
		decision.Participants[index].CompletionTime = player.CompletionTime
		decision.Participants[index].TargetValue = player.TargetValue
	}
	for _, playerID := range locked.WinningPlayerIDs {
		if index, ok := byID[playerID]; ok {
			decision.WinningPlayerRefs = append(decision.WinningPlayerRefs, decision.Participants[index].PlayerRef)
		}
	}
	return applied
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
