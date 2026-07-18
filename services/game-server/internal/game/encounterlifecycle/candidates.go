package encounterlifecycle

import "sort"

// CleanupCandidate contains only the authoritative facts needed for fallback
// population cleanup ordering.
type CleanupCandidate struct {
	EntityID              string
	NearestPlayerDistance float64
	Priority              Priority
}

// OrderCleanupCandidates returns a sorted copy. Candidates are ordered by
// farthest nearest-player distance, then lower retention priority, then stable
// entity ID.
func OrderCleanupCandidates(candidates []CleanupCandidate) []CleanupCandidate {
	ordered := append([]CleanupCandidate(nil), candidates...)
	sort.Slice(ordered, func(i, j int) bool {
		left, right := ordered[i], ordered[j]
		if left.NearestPlayerDistance != right.NearestPlayerDistance {
			return left.NearestPlayerDistance > right.NearestPlayerDistance
		}
		if left.Priority != right.Priority {
			return left.Priority < right.Priority
		}
		return left.EntityID < right.EntityID
	})
	return ordered
}
