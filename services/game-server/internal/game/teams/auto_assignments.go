package teams

import "fmt"

func ResolveAutoBalanced(participantIDs []string, teamCount int) (Assignments, error) {
	if teamCount < 2 || teamCount > len(OrderedIDs()) {
		return nil, fmt.Errorf("auto-balanced team count must be between 2 and %d", len(OrderedIDs()))
	}
	roster, err := CanonicalRoster(participantIDs)
	if err != nil {
		return nil, err
	}

	assignments := make(Assignments, len(roster))
	sizes := make([]int, teamCount)
	teamIDs := OrderedIDs()[:teamCount]
	for _, participantID := range roster {
		teamIndex := 0
		for index := 1; index < teamCount; index++ {
			if sizes[index] < sizes[teamIndex] {
				teamIndex = index
			}
		}
		assignments[participantID] = teamIDs[teamIndex]
		sizes[teamIndex]++
	}
	return assignments, nil
}
