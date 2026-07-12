package realtime

import "testing"

func TestRealtimeSessionStateIdentitySeparatesMatches(t *testing.T) {
	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 9, BaselineID: "baseline-9"})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, "projection")
	state.HotLaneCohorts.AsteroidRoutes["asteroid-1"] = HotUpdateRouteAsteroids
	state.HotLaneCohorts.BulletRoutes["bullet-1"] = HotUpdateRouteBullets
	state.HotLaneCohorts.AsteroidMode = HotLaneModeOverflow

	newMatch := NewRealtimeSessionState("player-1", "match-2")
	if newMatch.IdentityMatches(state.ReceiverID, state.MatchID) {
		t.Fatal("expected changed match identity to differ")
	}
	if _, ok := newMatch.LaneState(LaneWorld); ok || newMatch.LaneBaselineReady(LaneWorld) {
		t.Fatal("expected new match to start without lane or baseline state")
	}
	if projection, ok := newMatch.BaselineProjection(LaneWorld); ok || projection != nil {
		t.Fatal("expected new match to start without stored projections")
	}
	if len(newMatch.HotLaneCohorts.AsteroidRoutes) != 0 || len(newMatch.HotLaneCohorts.BulletRoutes) != 0 || newMatch.HotLaneCohorts.AsteroidMode != HotLaneModeInline || newMatch.HotLaneCohorts.BulletMode != HotLaneModeInline {
		t.Fatalf("expected fresh hot-lane cohort state, got %#v", newMatch.HotLaneCohorts)
	}
	if got := NextLaneSequence(RealtimeLaneState{}, false); got != 1 {
		t.Fatalf("expected new match sequence to restart at 1, got %d", got)
	}
}

func TestRealtimeSessionStateStartsUnsynced(t *testing.T) {
	state := NewRealtimeSessionState("player-1", "match-1")

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
	state := NewRealtimeSessionState("player-1", "match-1")

	if projection, ok := state.BaselineProjection(LaneWorld); ok || projection != nil {
		t.Fatalf("expected no baseline projection on new state, got %#v, %t", projection, ok)
	}
}

func TestRealtimeSessionStateStoresReadsAndClearsBaselineProjection(t *testing.T) {
	state := NewRealtimeSessionState("player-1", "match-1")

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
	state := NewRealtimeSessionState("player-1", "match-1")

	state.StoreBaselineProjection(LaneWorld, nil)

	if projection, ok := state.BaselineProjection(LaneWorld); ok || projection != nil {
		t.Fatalf("expected nil baseline projection to be ignored, got %#v, %t", projection, ok)
	}
}

func TestRealtimeSessionStateAcceptsFullLaneBaseline(t *testing.T) {
	state := NewRealtimeSessionState("player-1", "match-1")
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
	state := NewRealtimeSessionState("player-1", "match-1")
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
	state := NewRealtimeSessionState("player-1", "match-1")
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
	state := NewRealtimeSessionState("player-1", "match-1")
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
	first := NewRealtimeSessionState("player-1", "match-1")
	second := NewRealtimeSessionState("player-2", "match-1")

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
	state := NewRealtimeSessionState("player-1", "match-1")
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
	candidate := mustRealtimeLaneCandidate(WorldDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 12, BaselineID: "world-baseline", SnapshotID: "world-snapshot", SnapshotKind: SnapshotKind("delta"), IsFinalChunk: true}}, nil)
	metadata, ok := candidate.Metadata()
	if !ok {
		t.Fatal("expected world delta metadata to be returned")
	}
	if metadata.Lane != LaneWorld || metadata.Sequence != 12 || metadata.BaselineID != "world-baseline" || metadata.SnapshotID != "world-snapshot" || metadata.SnapshotKind != SnapshotKind("delta") {
		t.Fatalf("unexpected world delta metadata: %#v", metadata)
	}
}

func TestCandidateMetadataReturnsAsteroidHotDeltaMetadata(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{Type: PacketFamilyAsteroidDelta, Metadata: Metadata{Lane: LaneAsteroids, Sequence: 7, ServerSentMsec: 123, SnapshotKind: SnapshotKind("delta"), ChunkIndex: 1, ChunkCount: 3, IsFinalChunk: false}}, nil)
	metadata, ok := candidate.Metadata()
	if !ok {
		t.Fatal("expected asteroid hot delta metadata to be returned")
	}
	if metadata.Lane != LaneAsteroids || metadata.Sequence != 7 || metadata.ServerSentMsec != 123 || metadata.SnapshotKind != SnapshotKind("delta") || metadata.ChunkIndex != 1 || metadata.ChunkCount != 3 || metadata.IsFinalChunk {
		t.Fatalf("unexpected asteroid hot delta metadata: %#v", metadata)
	}
}

func TestCandidateMetadataReturnsBulletHotDeltaMetadata(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(BulletWireDeltaPacket{Type: PacketFamilyBulletDelta, Metadata: Metadata{Lane: LaneBullets, Sequence: 9, ServerSentMsec: 456, SnapshotKind: SnapshotKind("delta"), ChunkIndex: 2, ChunkCount: 4, IsFinalChunk: true}}, nil)
	metadata, ok := candidate.Metadata()
	if !ok {
		t.Fatal("expected bullet hot delta metadata to be returned")
	}
	if metadata.Lane != LaneBullets || metadata.Sequence != 9 || metadata.ServerSentMsec != 456 || metadata.SnapshotKind != SnapshotKind("delta") || metadata.ChunkIndex != 2 || metadata.ChunkCount != 4 || !metadata.IsFinalChunk {
		t.Fatalf("unexpected bullet hot delta metadata: %#v", metadata)
	}
}

func TestCandidateMetadataReturnsOverlayDeltaMetadata(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(OverlayLaneDelta{Metadata: Metadata{Lane: LaneOverlay, Sequence: 7, BaselineID: "overlay-baseline", SnapshotID: "overlay-snapshot", SnapshotKind: SnapshotKind("delta"), IsFinalChunk: true}}, nil)
	metadata, ok := candidate.Metadata()
	if !ok {
		t.Fatal("expected overlay delta metadata to be returned")
	}
	if metadata.Lane != LaneOverlay || metadata.Sequence != 7 || metadata.BaselineID != "overlay-baseline" || metadata.SnapshotID != "overlay-snapshot" || metadata.SnapshotKind != SnapshotKind("delta") {
		t.Fatalf("unexpected overlay delta metadata: %#v", metadata)
	}
}

func TestCandidateMetadataReturnsSessionDeltaMetadata(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(SessionLaneDelta{Metadata: Metadata{Lane: LaneSession, Sequence: 5, BaselineID: "session-baseline", SnapshotID: "session-snapshot", SnapshotKind: SnapshotKind("delta"), IsFinalChunk: true}}, nil)
	metadata, ok := candidate.Metadata()
	if !ok {
		t.Fatal("expected session delta metadata to be returned")
	}
	if metadata.Lane != LaneSession || metadata.Sequence != 5 || metadata.BaselineID != "session-baseline" || metadata.SnapshotID != "session-snapshot" || metadata.SnapshotKind != SnapshotKind("delta") {
		t.Fatalf("unexpected session delta metadata: %#v", metadata)
	}
}

func TestAdvanceMetadataForSuccessfulWriteAdvancesEventLaneSequence(t *testing.T) {
	state := NewRealtimeSessionState("player-1", "match-1")
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

func TestRequireFullBaselineOnlyInvalidatesRequestedLane(t *testing.T) {
	state := NewRealtimeSessionState("player-1", "match-1")
	world := Metadata{Lane: LaneWorld, Sequence: 7, SnapshotID: "world-snapshot", BaselineID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true}
	overlay := Metadata{Lane: LaneOverlay, Sequence: 9, SnapshotID: "overlay-snapshot", BaselineID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true}
	state.UpdateLane(LaneWorld, world)
	state.UpdateLane(LaneOverlay, overlay)
	state.MarkBaselineReady(LaneWorld)
	state.MarkBaselineReady(LaneOverlay)
	state.StoreBaselineProjection(LaneWorld, "world-projection")
	state.StoreBaselineProjection(LaneOverlay, "overlay-projection")
	if !state.RequireFullBaseline(LaneWorld) {
		t.Fatal("expected invalidation")
	}
	gotWorld, _ := state.LaneState(LaneWorld)
	if gotWorld.Metadata() != world || state.LaneBaselineReady(LaneWorld) {
		t.Fatalf("world changed: %#v", gotWorld)
	}
	if projection, ok := state.BaselineProjection(LaneWorld); ok || projection != nil {
		t.Fatalf("world projection not cleared: %#v, %t", projection, ok)
	}
	gotOverlay, _ := state.LaneState(LaneOverlay)
	if gotOverlay.Metadata() != overlay || !state.LaneBaselineReady(LaneOverlay) {
		t.Fatalf("overlay changed: %#v", gotOverlay)
	}
	if projection, ok := state.BaselineProjection(LaneOverlay); !ok || projection != "overlay-projection" {
		t.Fatalf("overlay projection changed: %#v, %t", projection, ok)
	}
	if state.RequireFullBaseline(LaneControl) {
		t.Fatal("expected unsupported lane rejection")
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
