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
