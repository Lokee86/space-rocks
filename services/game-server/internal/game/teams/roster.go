package teams

import (
	"fmt"
	"sort"
)

func CanonicalRoster(participantIDs []string) ([]string, error) {
	roster := append([]string(nil), participantIDs...)
	seen := make(map[string]struct{}, len(roster))
	for _, participantID := range roster {
		if participantID == "" {
			return nil, fmt.Errorf("participant ID cannot be blank")
		}
		if _, exists := seen[participantID]; exists {
			return nil, fmt.Errorf("participant ID %q is duplicated", participantID)
		}
		seen[participantID] = struct{}{}
	}
	sort.Strings(roster)
	return roster, nil
}
