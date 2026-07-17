package teams

func ResolveFFA(participantIDs []string) (Assignments, error) {
	return resolveBaseline(participantIDs, NoTeam)
}

func ResolveCoop(participantIDs []string) (Assignments, error) {
	return resolveBaseline(participantIDs, Team1)
}

func ResolveCoOp(participantIDs []string) (Assignments, error) {
	return ResolveCoop(participantIDs)
}

func resolveBaseline(participantIDs []string, teamID ID) (Assignments, error) {
	roster, err := CanonicalRoster(participantIDs)
	if err != nil {
		return nil, err
	}
	assignments := make(Assignments, len(roster))
	for _, participantID := range roster {
		assignments[participantID] = teamID
	}
	return assignments, nil
}
