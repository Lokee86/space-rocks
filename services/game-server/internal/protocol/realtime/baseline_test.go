package realtime

import "testing"

func TestRealtimeSessionStateStartsUnsynced(t *testing.T) {
	state := NewRealtimeSessionState("player-1")

	if state.ReceiverID != "player-1" {
		t.Fatalf("expected receiver ID to be preserved, got %q", state.ReceiverID)
	}
	if _, ok := state.LaneState(LaneWorld); ok {
		t.Fatal("expected lane to start unsynced")
	}
}

func TestNextLaneSequenceReturnsOneForUnsyncedLane(t *testing.T) {
	if got := NextLaneSequence(RealtimeLaneState{Sequence: 7}, false); got != 1 {
		t.Fatalf("NextLaneSequence(unsynced) = %d, want 1", got)
	}
}

func TestNextLaneSequenceReturnsOneForZeroSequence(t *testing.T) {
	if got := NextLaneSequence(RealtimeLaneState{Sequence: 0}, true); got != 1 {
		t.Fatalf("NextLaneSequence(zero) = %d, want 1", got)
	}
}

func TestNextLaneSequenceReturnsTwoForSequenceOne(t *testing.T) {
	if got := NextLaneSequence(RealtimeLaneState{Sequence: 1}, true); got != 2 {
		t.Fatalf("NextLaneSequence(1) = %d, want 2", got)
	}
}

func TestNextLaneSequenceReturnsEightForSequenceSeven(t *testing.T) {
	if got := NextLaneSequence(RealtimeLaneState{Sequence: 7}, true); got != 8 {
		t.Fatalf("NextLaneSequence(7) = %d, want 8", got)
	}
}

func TestRealtimeSessionStateStartsWithNoBaselineProjection(t *testing.T) {
	state := NewRealtimeSessionState("player-1")

	if projection, ok := state.BaselineProjection(LaneWorld); ok || projection != nil {
		t.Fatalf("expected no baseline projection on new state, got %#v, %t", projection, ok)
	}
}

func TestRealtimeSessionStateStoresReadsAndClearsBaselineProjection(t *testing.T) {
	state := NewRealtimeSessionState("player-1")

	state.StoreBaselineProjection(LaneWorld, "world-projection")

	projection, ok := state.BaselineProjection(LaneWorld)
	if !ok {
		t.Fatal("expected baseline projection to be stored")
	}
	if projection != "world-projection" {
		t.Fatalf("expected stored projection to be returned, got %#v", projection)
	}

	state.ClearBaselineProjection(LaneWorld)

	projection, ok = state.BaselineProjection(LaneWorld)
	if ok || projection != nil {
		t.Fatalf("expected baseline projection to be cleared, got %#v, %t", projection, ok)
	}
}

func TestRealtimeSessionStateIgnoresNilBaselineProjection(t *testing.T) {
	state := NewRealtimeSessionState("player-1")

	state.StoreBaselineProjection(LaneWorld, nil)

	if projection, ok := state.BaselineProjection(LaneWorld); ok || projection != nil {
		t.Fatalf("expected nil baseline projection to be ignored, got %#v, %t", projection, ok)
	}
}

func TestRealtimeSessionStateAcceptsFullLaneBaseline(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{
		Lane:           LaneWorld,
		Sequence:       4,
		SnapshotID:     "snapshot-1",
		BaselineID:     "baseline-1",
		SnapshotKind:   SnapshotKind("full"),
		ChunkIndex:     0,
		ChunkCount:     1,
		IsFinalChunk:   true,
		ServerSentMsec: 123,
	})

	laneState, ok := state.LaneState(LaneWorld)
	if !ok {
		t.Fatal("expected lane state to be present after accepting baseline")
	}
	if laneState.BaselineID != "baseline-1" || laneState.SnapshotID != "snapshot-1" || laneState.Sequence != 4 || laneState.IsFinalChunk != true {
		t.Fatalf("expected accepted baseline metadata to be tracked, got %#v", laneState)
	}
}

func TestRealtimeSessionStateChunkedBaselineRemainsIncompleteUntilFinalChunk(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneSession, Metadata{
		Lane:           LaneSession,
		Sequence:       8,
		SnapshotID:     "snapshot-2",
		BaselineID:     "baseline-2",
		SnapshotKind:   SnapshotKind("full"),
		ChunkIndex:     0,
		ChunkCount:     2,
		IsFinalChunk:   false,
		ServerSentMsec: 456,
	})

	laneState, ok := state.LaneState(LaneSession)
	if !ok {
		t.Fatal("expected lane state to be present after first chunk")
	}
	if laneState.IsFinalChunk {
		t.Fatal("expected chunked baseline to remain incomplete until final chunk")
	}

	state.UpdateLane(LaneSession, Metadata{
		Lane:           LaneSession,
		Sequence:       8,
		SnapshotID:     "snapshot-2",
		BaselineID:     "baseline-2",
		SnapshotKind:   SnapshotKind("full"),
		ChunkIndex:     1,
		ChunkCount:     2,
		IsFinalChunk:   true,
		ServerSentMsec: 456,
	})

	laneState, ok = state.LaneState(LaneSession)
	if !ok || !laneState.IsFinalChunk {
		t.Fatalf("expected final chunk to complete baseline, got %#v", laneState)
	}
}

func TestRealtimeSessionStateTracksBaselinePerLane(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, SnapshotID: "world-snapshot", BaselineID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 2, SnapshotID: "overlay-snapshot", BaselineID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})

	worldState, ok := state.LaneState(LaneWorld)
	if !ok || worldState.BaselineID != "world-baseline" || worldState.SnapshotID != "world-snapshot" {
		t.Fatalf("expected world lane baseline to be tracked, got %#v", worldState)
	}
	overlayState, ok := state.LaneState(LaneOverlay)
	if !ok || overlayState.BaselineID != "overlay-baseline" || overlayState.SnapshotID != "overlay-snapshot" {
		t.Fatalf("expected overlay lane baseline to be tracked, got %#v", overlayState)
	}
}

func TestDecideResyncDetectsWrongAndMissingBaselines(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	observed := RealtimeLaneState{
		Lane:       LaneWorld,
		Sequence:   9,
		BaselineID: "baseline-actual",
		SnapshotID: "snapshot-actual",
	}

	wrong := DecideResync(state, LaneWorld, "baseline-expected", "", observed, true)
	if wrong.Kind != ResyncDecisionWrongBaseline || wrong.BaselineID != "baseline-actual" || wrong.SnapshotID != "snapshot-actual" || wrong.Sequence != 9 {
		t.Fatalf("expected wrong baseline decision, got %#v", wrong)
	}

	missing := DecideResync(state, LaneSession, "", "baseline-required", RealtimeLaneState{}, false)
	if missing.Kind != ResyncDecisionMissingBaseline || missing.BaselineID != "baseline-required" {
		t.Fatalf("expected missing baseline decision, got %#v", missing)
	}
}

func TestOverlayBaselineIsReceiverSpecific(t *testing.T) {
	first := NewRealtimeSessionState("player-1")
	second := NewRealtimeSessionState("player-2")

	first.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, SnapshotID: "overlay-1", BaselineID: "baseline-1", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	second.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, SnapshotID: "overlay-2", BaselineID: "baseline-2", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})

	firstLane, ok := first.LaneState(LaneOverlay)
	if !ok || firstLane.BaselineID != "baseline-1" {
		t.Fatalf("expected receiver-specific overlay baseline for player-1, got %#v", firstLane)
	}
	secondLane, ok := second.LaneState(LaneOverlay)
	if !ok || secondLane.BaselineID != "baseline-2" {
		t.Fatalf("expected receiver-specific overlay baseline for player-2, got %#v", secondLane)
	}
}


func TestRealtimeSessionStateIgnoresStaleSequencesAndTracksWrongBaselineResync(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 10, SnapshotID: "snapshot-new", BaselineID: "baseline-new", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 9, SnapshotID: "snapshot-old", BaselineID: "baseline-old", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})

	laneState, ok := state.LaneState(LaneWorld)
	if !ok || laneState.Sequence != 10 || laneState.BaselineID != "baseline-new" {
		t.Fatalf("expected stale sequence to be ignored, got %#v", laneState)
	}

	decision := DecideResync(state, LaneWorld, "baseline-wrong", "baseline-required", laneState, true)
	if decision.Kind != ResyncDecisionWrongBaseline {
		t.Fatalf("expected wrong baseline resync decision, got %#v", decision)
	}
}

func TestCandidateMetadataReturnsWorldDeltaMetadata(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	candidate := RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindDelta, Delta: WorldDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 12, BaselineID: "world-baseline", SnapshotID: "world-snapshot", SnapshotKind: SnapshotKind("delta"), IsFinalChunk: true}}}

	metadata, ok := CandidateMetadata(candidate, state)
	if !ok {
		t.Fatal("expected world delta metadata to be returned")
	}
	if metadata.Lane != LaneWorld || metadata.Sequence != 12 || metadata.BaselineID != "world-baseline" || metadata.SnapshotID != "world-snapshot" || metadata.SnapshotKind != SnapshotKind("delta") {
		t.Fatalf("unexpected world delta metadata: %#v", metadata)
	}
}

func TestCandidateMetadataReturnsOverlayDeltaMetadata(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	candidate := RealtimeLaneCandidate{Lane: LaneOverlay, Kind: RealtimeLaneCandidateKindDelta, Delta: OverlayLaneDelta{Metadata: Metadata{Lane: LaneOverlay, Sequence: 7, BaselineID: "overlay-baseline", SnapshotID: "overlay-snapshot", SnapshotKind: SnapshotKind("delta"), IsFinalChunk: true}}}

	metadata, ok := CandidateMetadata(candidate, state)
	if !ok {
		t.Fatal("expected overlay delta metadata to be returned")
	}
	if metadata.Lane != LaneOverlay || metadata.Sequence != 7 || metadata.BaselineID != "overlay-baseline" || metadata.SnapshotID != "overlay-snapshot" || metadata.SnapshotKind != SnapshotKind("delta") {
		t.Fatalf("unexpected overlay delta metadata: %#v", metadata)
	}
}

func TestCandidateMetadataReturnsSessionDeltaMetadata(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	candidate := RealtimeLaneCandidate{Lane: LaneSession, Kind: RealtimeLaneCandidateKindDelta, Delta: SessionLaneDelta{Metadata: Metadata{Lane: LaneSession, Sequence: 5, BaselineID: "session-baseline", SnapshotID: "session-snapshot", SnapshotKind: SnapshotKind("delta"), IsFinalChunk: true}}}

	metadata, ok := CandidateMetadata(candidate, state)
	if !ok {
		t.Fatal("expected session delta metadata to be returned")
	}
	if metadata.Lane != LaneSession || metadata.Sequence != 5 || metadata.BaselineID != "session-baseline" || metadata.SnapshotID != "session-snapshot" || metadata.SnapshotKind != SnapshotKind("delta") {
		t.Fatalf("unexpected session delta metadata: %#v", metadata)
	}
}

func TestCandidateMetadataFallsBackToLaneStateForUnsupportedDeltaShape(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 22, BaselineID: "world-baseline", SnapshotID: "world-snapshot", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	candidate := RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindDelta}

	metadata, ok := CandidateMetadata(candidate, state)
	if !ok {
		t.Fatal("expected fallback metadata to be returned")
	}
	if metadata.Sequence != 22 || metadata.BaselineID != "world-baseline" || metadata.SnapshotID != "world-snapshot" || metadata.SnapshotKind != SnapshotKind("full") {
		t.Fatalf("unexpected fallback metadata: %#v", metadata)
	}
}

func TestAdvanceMetadataForSuccessfulWriteAdvancesEventLaneSequence(t *testing.T) {
	state := NewRealtimeSessionState("player-1")
	metadata := Metadata{
		Lane:         LaneEvent,
		Sequence:     0,
		SnapshotID:   "event-batch-0",
		SnapshotKind: SnapshotKind("batch"),
		ChunkIndex:   0,
		ChunkCount:   1,
		IsFinalChunk: true,
	}

	state.UpdateLane(LaneEvent, AdvanceMetadataForSuccessfulWrite(LaneEvent, metadata))

	laneState, ok := state.LaneState(LaneEvent)
	if !ok {
		t.Fatal("expected event lane state after successful write metadata persists")
	}
	if laneState.Sequence != 1 {
		t.Fatalf("event lane sequence = %d, want 1", laneState.Sequence)
	}
	if laneState.SnapshotID != "event-batch-1" {
		t.Fatalf("event lane snapshot id = %q, want event-batch-1", laneState.SnapshotID)
	}
}


func TestFullBaselineID(t *testing.T) {
	tests := []struct {
		name     string
		lane     Lane
		sequence int
		want     string
	}{
		{name: "world", lane: LaneWorld, sequence: 9, want: "world-baseline-9"},
		{name: "overlay", lane: LaneOverlay, sequence: 4, want: "overlay-baseline-4"},
		{name: "session", lane: LaneSession, sequence: 5, want: "session-baseline-5"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := FullBaselineID(tc.lane, tc.sequence); got != tc.want {
				t.Fatalf("FullBaselineID(%q, %d) = %q, want %q", tc.lane, tc.sequence, got, tc.want)
			}
		})
	}
}

func TestDeltaSnapshotID(t *testing.T) {
	tests := []struct {
		name     string
		lane     Lane
		sequence int
		want     string
	}{
		{name: "world", lane: LaneWorld, sequence: 10, want: "world-snapshot-10"},
		{name: "overlay", lane: LaneOverlay, sequence: 11, want: "overlay-snapshot-11"},
		{name: "session", lane: LaneSession, sequence: 12, want: "session-snapshot-12"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := DeltaSnapshotID(tc.lane, tc.sequence); got != tc.want {
				t.Fatalf("DeltaSnapshotID(%q, %d) = %q, want %q", tc.lane, tc.sequence, got, tc.want)
			}
		})
	}
}
