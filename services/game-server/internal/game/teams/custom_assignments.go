package teams

import "fmt"

func ResolveCustom(participantIDs []string, requested Assignments) (Assignments, error) {
	roster, err := CanonicalRoster(participantIDs)
	if err != nil {
		return nil, err
	}
	if len(requested) != len(roster) {
		return nil, fmt.Errorf("custom assignments must include every participant exactly once")
	}

	assignments := make(Assignments, len(roster))
	for _, participantID := range roster {
		teamID, ok := requested[participantID]
		if !ok {
			return nil, fmt.Errorf("custom assignment is missing participant %q", participantID)
		}
		if err := ValidateTeamID(teamID); err != nil {
			return nil, fmt.Errorf("custom assignment for participant %q: %w", participantID, err)
		}
		if teamID == NoTeam {
			return nil, fmt.Errorf("custom assignment for participant %q cannot use NoTeam", participantID)
		}
		assignments[participantID] = teamID
	}
	for participantID := range requested {
		if _, ok := assignments[participantID]; !ok {
			return nil, fmt.Errorf("custom assignment references unknown participant %q", participantID)
		}
	}
	return assignments, nil
}
