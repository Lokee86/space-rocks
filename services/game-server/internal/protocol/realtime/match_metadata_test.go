package realtime

import "testing"

func TestMatchMetadataIsPreservedAcrossRealtimePacketFamilies(t *testing.T) {
	tests := []struct {
		name string
		lane Lane
		kind RealtimeLaneCandidateKind
	}{
		{name: "world_full", lane: LaneWorld, kind: RealtimeLaneCandidateKindFull},
		{name: "world_delta", lane: LaneWorld, kind: RealtimeLaneCandidateKindDelta},
		{name: "overlay_full", lane: LaneOverlay, kind: RealtimeLaneCandidateKindFull},
		{name: "overlay_delta", lane: LaneOverlay, kind: RealtimeLaneCandidateKindDelta},
		{name: "session_full", lane: LaneSession, kind: RealtimeLaneCandidateKindFull},
		{name: "session_delta", lane: LaneSession, kind: RealtimeLaneCandidateKindDelta},
		{name: "asteroid_delta", lane: LaneAsteroids, kind: RealtimeLaneCandidateKindDelta},
		{name: "bullet_delta", lane: LaneBullets, kind: RealtimeLaneCandidateKindDelta},
		{name: "asteroids_lifecycle", lane: LaneAsteroidsLifecycle, kind: RealtimeLaneCandidateKindDelta},
		{name: "bullets_lifecycle", lane: LaneBulletsLifecycle, kind: RealtimeLaneCandidateKindDelta},
		{name: "event_batch", lane: LaneEvent, kind: RealtimeLaneCandidateKindEventBatch},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			candidate := testCandidate(tc.lane, tc.kind)
			candidate.MatchID = "match-1"

			readable, err := WireLanePacket(candidate)
			if err != nil {
				t.Fatalf("WireLanePacket() error = %v", err)
			}
			if got := readable["match_id"]; got != "match-1" {
				t.Fatalf("readable match_id = %#v, want %q", got, "match-1")
			}

			compact := CompactWirePacket(readable)
			if got := compact["mid"]; got != "match-1" {
				t.Fatalf("compact mid = %#v, want %q", got, "match-1")
			}

			metadata, ok := candidate.Metadata()
			if !ok {
				t.Fatal("candidate.Metadata() reported no metadata")
			}
			if metadata.MatchID != "match-1" {
				t.Fatalf("candidate metadata MatchID = %q, want %q", metadata.MatchID, "match-1")
			}
		})
	}
}
