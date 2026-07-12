package realtime

import "testing"

func mustExpandRealtimeCandidateChunks(candidates []RealtimeLaneCandidate) []RealtimeLaneCandidate {
	return mustExpandRealtimeCandidateChunksT(nil, candidates)
}

func mustExpandRealtimeCandidateChunksT(t *testing.T, candidates []RealtimeLaneCandidate) []RealtimeLaneCandidate {
	expanded, err := ExpandRealtimeCandidateChunks(candidates)
	if err != nil {
		if t != nil {
			t.Fatalf("expand realtime candidate chunks: %v", err)
		}
		panic(err)
	}
	return expanded
}
