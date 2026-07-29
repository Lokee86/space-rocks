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
	for index, fact := range participants {
		disposition := fact.Disposition
		if disposition == "" {
			disposition = DispositionParticipated
		}
		results[index] = ParticipantResult{
			PlayerRef:   fact.PlayerRef,
			TeamID:      fact.TeamID,
			Disposition: disposition,
			FinalScore:  fact.Score,
			ShipDeaths:  fact.ShipDeaths,
		}
	}

	switch input.LockedDecision.TerminalStatus {
	case rules.TerminalCompleted, rules.TerminalFailed, rules.TerminalCancelled,
		rules.TerminalInvalid, rules.TerminalAdministrativelyTerminated:
	default:
		return MatchDecision{}, fmt.Errorf("locked match decision has invalid terminal status %q", input.LockedDecision.TerminalStatus)
	}
	decision := MatchDecision{
		TerminalStatus: TerminalStatus(input.LockedDecision.TerminalStatus),
		EndReason:      input.LockedDecision.EndReason,
		Participants:   results,
	}
	if decision.EndReason == "" {
		decision.EndReason = input.EndReason
	}
	if decision.EndReason == "" {
		return MatchDecision{}, fmt.Errorf("locked match decision requires end reason")
	}
	if err := applyLockedParticipantDecision(&decision, input.LockedDecision); err != nil {
		return MatchDecision{}, err
	}
	decision.Teams, decision.WinningTeamRefs = resolveLockedTeamOutcomes(input.TeamStructure, decision.Participants)
	return decision, nil
}

func applyLockedParticipantDecision(decision *MatchDecision, locked rules.MatchDecision) error {
	byID := make(map[string]int, len(decision.Participants))
	for index, participant := range decision.Participants {
		byID[participant.PlayerRef.GamePlayerID] = index
	}

	seen := make(map[string]struct{}, len(locked.Players))
	for _, player := range locked.Players {
		if _, duplicate := seen[player.ID]; duplicate {
			return fmt.Errorf("locked match decision contains duplicate player %q", player.ID)
		}
		seen[player.ID] = struct{}{}
		index, ok := byID[player.ID]
		if !ok {
			return fmt.Errorf("locked match decision references unknown player %q", player.ID)
		}
		switch player.Outcome {
		case rules.OutcomeWon, rules.OutcomeLost, rules.OutcomeDraw,
			rules.OutcomeCompleted, rules.OutcomeFailed, rules.OutcomeAborted:
		default:
			return fmt.Errorf("locked match decision has invalid outcome %q for player %q", player.Outcome, player.ID)
		}
		decision.Participants[index].Outcome = Outcome(player.Outcome)
		decision.Participants[index].Placement = player.Placement
		decision.Participants[index].CompletionTime = player.CompletionTime
		decision.Participants[index].TargetValue = player.TargetValue
	}
	for playerID := range byID {
		if _, ok := seen[playerID]; !ok {
			return fmt.Errorf("locked match decision is missing player %q", playerID)
		}
	}

	winners := make(map[string]struct{}, len(locked.WinningPlayerIDs))
	for _, playerID := range locked.WinningPlayerIDs {
		if _, duplicate := winners[playerID]; duplicate {
			return fmt.Errorf("locked match decision contains duplicate winner %q", playerID)
		}
		winners[playerID] = struct{}{}
		index, ok := byID[playerID]
		if !ok {
			return fmt.Errorf("locked match decision winner %q is not a participant", playerID)
		}
		if decision.Participants[index].Outcome != OutcomeWon {
			return fmt.Errorf("locked match decision winner %q does not have a won outcome", playerID)
		}
		decision.WinningPlayerRefs = append(decision.WinningPlayerRefs, decision.Participants[index].PlayerRef)
	}
	return nil
}
