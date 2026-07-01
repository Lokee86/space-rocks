package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestActiveLaneMetricsRecordBytesAndCounts(t *testing.T) {
	summary := SendPlanSummary{
		IncludedCount:   2,
		DeferredCount:   1,
		SupersededCount: 0,
		RequiredCount:   2,
		CreateCount:     1,
		UpdateCount:     1,
		DeleteCount:     0,
	}

	result := ActiveRealtimeResult{
		SelectedCandidates: []RealtimeLaneCandidate{{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull}},
		SendPlan:           SendPlan{Summary: summary},
		EncodedBytes:       map[Lane]int{LaneWorld: 128},
		Mode:               "active",
	}

	records := ActiveLaneMetricRecords(result)
	if len(records) != 1 {
		t.Fatalf("expected 1 metric record, got %d", len(records))
	}
	if records[0].Bytes != 128 {
		t.Fatalf("active metric bytes = %d, want 128", records[0].Bytes)
	}
}

func TestCandidateProjectionReturnsProjectionForFullCandidate(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane:       LaneWorld,
		Kind:       RealtimeLaneCandidateKindFull,
		Full:       "wire-full",
		Projection: "baseline-projection",
	}

	projection, ok := CandidateProjection(candidate)
	if !ok {
		t.Fatal("expected full candidate projection to be returned")
	}
	if projection != "baseline-projection" {
		t.Fatalf("expected projection to match stored value, got %#v", projection)
	}
}

func TestCandidateProjectionReturnsFalseForNilProjection(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneOverlay,
		Kind: RealtimeLaneCandidateKindFull,
		Full: "wire-full",
	}

	projection, ok := CandidateProjection(candidate)
	if ok || projection != nil {
		t.Fatalf("expected nil projection to be ignored, got %#v, %t", projection, ok)
	}
}

func TestCandidateProjectionReturnsFalseForEventBatchCandidate(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane:       LaneEvent,
		Kind:       RealtimeLaneCandidateKindEventBatch,
		Full:       "event-batch",
		Projection: "should-be-ignored",
	}

	projection, ok := CandidateProjection(candidate)
	if ok || projection != nil {
		t.Fatalf("expected event-batch candidate to have no projection, got %#v, %t", projection, ok)
	}
}

func TestBuildActiveRealtimeResultEncodesOnlyEnvelopePackets(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:         "player-1",
		Lives:          3,
		ServerSentMsec: 1234,
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing", X: 1, Y: 2, Rotation: 3, Health: 4, Shields: 5},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 9, Lives: 3, RespawnCooldown: 1.25, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite", SecondaryWeaponID: "mine", SecondaryAmmoPolicy: "limited"},
		},
		PlayerLifecycle: map[string]string{"player-1": "active"},
		Bullets: map[string]runtime.BulletState{
			"bullet-1": {ID: "bullet-1", OwnerID: "player-1", X: 6, Y: 7, Rotation: 8, WeaponID: "laser", ProjectileType: "bolt"},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-1": {ID: "asteroid-1", X: 9, Y: 10, Size: 2, Health: 11, Scale: 1.5, Variant: 3},
		},
		Pickups: map[string]runtime.PickupState{
			"pickup-1": {ID: "pickup-1", Type: "shield", PickupClass: "armor", X: 12, Y: 13, Health: 1, AgeSeconds: 4.5, LifespanSeconds: 9.5},
		},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.UpdateLane(LaneSession, Metadata{Lane: LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)

	result := BuildActiveRealtimeResult(snapshot, state)
	if len(result.Candidates) == 0 {
		t.Fatal("expected active realtime result to emit candidates")
	}
	if len(result.SelectedCandidates) != len(result.Candidates) {
		t.Fatalf("expected selected candidates to match candidates in this baseline case, got %d selected and %d candidates", len(result.SelectedCandidates), len(result.Candidates))
	}

	for _, candidate := range result.SelectedCandidates {
		encodedPacket, ok := result.EncodedPackets[candidate.Lane]
		if !ok || len(encodedPacket) == 0 {
			t.Fatalf("expected encoded packet for lane=%q kind=%q", candidate.Lane, candidate.Kind)
		}
		wire := mustDecodeWirePacket(t, encodedPacket)
		if gotType, ok := wire["type"].(string); !ok || gotType == "" {
			t.Fatalf("expected non-empty top-level type for lane=%q kind=%q, got %#v", candidate.Lane, candidate.Kind, wire)
		}
		if gotLane, ok := wire["lane"].(string); !ok || gotLane == "" {
			t.Fatalf("expected non-empty top-level lane for lane=%q kind=%q, got %#v", candidate.Lane, candidate.Kind, wire)
		}
		assertNotNakedLanePayload(t, candidate.Lane, wire)
	}
}

func TestBuildActiveRealtimeResultSelectsFullPacketsWithoutStoredBaselines(t *testing.T) {
	snapshot := tinyActiveBoundarySnapshot()
	result := BuildActiveRealtimeResult(snapshot, NewRealtimeSessionState("player-1"))

	assertSelectedCandidate(t, result, LaneWorld, RealtimeLaneCandidateKindFull)
	assertSelectedCandidate(t, result, LaneOverlay, RealtimeLaneCandidateKindFull)
	assertSelectedCandidate(t, result, LaneSession, RealtimeLaneCandidateKindFull)

	assertEncodedPacketTypeAndLane(t, result, LaneWorld, PacketFamilyWorldFull, string(LaneWorld))
	assertEncodedPacketTypeAndLane(t, result, LaneOverlay, PacketFamilyOverlayFull, string(LaneOverlay))
	assertEncodedPacketTypeAndLane(t, result, LaneSession, PacketFamilySessionFull, string(LaneSession))
}

func TestBuildActiveRealtimeResultEmitsNoWorldOverlayOrSessionPacketsWhenStoredBaselinesMatch(t *testing.T) {
	snapshot := tinyActiveBoundarySnapshot()
	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, snapshot, 1))
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.StoreBaselineProjection(LaneOverlay, mustOverlayWireFull(t, snapshot, "player-1", 1))
	state.UpdateLane(LaneSession, Metadata{Lane: LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.StoreBaselineProjection(LaneSession, mustSessionWireFull(t, snapshot, 1))

	result := BuildActiveRealtimeResult(snapshot, state)
	assertNoSelectedCandidate(t, result, LaneWorld)
	assertNoSelectedCandidate(t, result, LaneOverlay)
	assertNoSelectedCandidate(t, result, LaneSession)
}

func TestBuildActiveRealtimeResultSelectsDeltaPacketsForChangedStoredBaselines(t *testing.T) {
	snapshot := tinyActiveBoundarySnapshot()
	snapshot.Players["player-1"] = runtime.ShipState{ID: "player-1", ShipType: "v_wing", X: 2, Y: 1, Rotation: 0, Health: 5, Shields: 0}
	snapshot.PlayerSessions["player-1"] = game.PlayerSessionState{ID: "player-1", ShipType: "v_wing", Score: 7, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"}
	snapshot.PlayerLifecycle["player-1"] = "active"
	snapshot.TotalAsteroids = 1

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 2, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, game.GameplayPresentationSnapshot{SelfID: "player-1", Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing", X: 1, Y: 1, Rotation: 0, Health: 5, Shields: 0}}}, 1))

	result := BuildActiveRealtimeResult(snapshot, state)
	assertSelectedCandidate(t, result, LaneWorld, RealtimeLaneCandidateKindDelta)
	assertEncodedPacketTypeAndLane(t, result, LaneWorld, PacketTypeWorldDelta, string(LaneWorld))
}

func tinyActiveBoundarySnapshot() game.GameplayPresentationSnapshot {
	return game.GameplayPresentationSnapshot{
		SelfID:         "player-1",
		Lives:          3,
		ServerSentMsec: 1234,
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing", X: 1, Y: 1, Rotation: 0, Health: 5, Shields: 0},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 5, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"},
		},
		PlayerLifecycle: map[string]string{"player-1": "active"},
	}
}

func assertSelectedCandidate(t *testing.T, result ActiveRealtimeResult, lane Lane, kind RealtimeLaneCandidateKind) {
	t.Helper()
	for _, candidate := range result.SelectedCandidates {
		if candidate.Lane == lane && candidate.Kind == kind {
			return
		}
	}
	t.Fatalf("expected selected candidate lane=%q kind=%q, got %#v", lane, kind, result.SelectedCandidates)
}

func assertNoSelectedCandidate(t *testing.T, result ActiveRealtimeResult, lane Lane) {
	t.Helper()
	for _, candidate := range result.SelectedCandidates {
		if candidate.Lane == lane {
			t.Fatalf("expected no selected candidate for lane=%q, got %#v", lane, result.SelectedCandidates)
		}
	}
	if _, ok := result.EncodedPackets[lane]; ok {
		t.Fatalf("expected no encoded packet for lane=%q, got %#v", lane, result.EncodedPackets[lane])
	}
}

func assertEncodedPacketTypeAndLane(t *testing.T, result ActiveRealtimeResult, lane Lane, wantType string, wantLane string) {
	t.Helper()
	encoded, ok := result.EncodedPackets[lane]
	if !ok || len(encoded) == 0 {
		t.Fatalf("expected encoded packet for lane=%q", lane)
	}
	wire := mustDecodeWirePacket(t, encoded)
	if got, ok := wire["type"].(string); !ok || got != wantType {
		t.Fatalf("expected type=%q for lane=%q, got %#v", wantType, lane, wire)
	}
	if got, ok := wire["lane"].(string); !ok || got != wantLane {
		t.Fatalf("expected lane=%q for lane=%q, got %#v", wantLane, lane, wire)
	}
}

func TestIncludedRealtimeLaneCandidatesSkipsDeferredRecordsInOrder(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull},
		{Lane: LaneOverlay, Kind: RealtimeLaneCandidateKindFull},
		{Lane: LaneSession, Kind: RealtimeLaneCandidateKindEventBatch},
	}
	included := []ScheduleRecord{
		{CandidateIndex: 2},
		{CandidateIndex: 0},
	}
	deferred := []ScheduleRecord{
		{CandidateIndex: 1},
	}

	if len(deferred) != 1 {
		t.Fatalf("expected 1 deferred record, got %d", len(deferred))
	}

	selected := IncludedRealtimeLaneCandidates(candidates, included)
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected candidates, got %d", len(selected))
	}
	if selected[0].Lane != LaneSession || selected[1].Lane != LaneWorld {
		t.Fatalf("selected candidates = %#v, want session then world", selected)
	}
}

func TestBuildActiveRealtimeResultUsesSelectedCandidatesOnly(t *testing.T) {
	result := ActiveRealtimeResult{
		SelectedCandidates: []RealtimeLaneCandidate{
			{Lane: LaneOverlay, Kind: RealtimeLaneCandidateKindFull},
		},
		SendPlan: SendPlan{
			Summary: SendPlanSummary{IncludedCount: 1},
		},
		EncodedPackets: map[Lane][]byte{
			LaneOverlay: []byte(`{"type":"overlay_full","lane":"overlay"}`),
		},
		EncodedBytes: map[Lane]int{
			LaneOverlay: 42,
		},
	}

	records := ActiveLaneMetricRecords(result)
	if len(records) != 1 {
		t.Fatalf("expected 1 metric record, got %d", len(records))
	}
	if records[0].Lane != LaneOverlay {
		t.Fatalf("expected metric record for overlay, got lane=%q", records[0].Lane)
	}
	if _, ok := result.EncodedPackets[LaneWorld]; ok {
		t.Fatal("expected world packet to be absent when not selected")
	}
	if _, ok := result.EncodedPackets[LaneOverlay]; !ok {
		t.Fatal("expected overlay packet to be present when selected")
	}
}

func assertNotNakedLanePayload(t *testing.T, lane Lane, wire map[string]any) {
	t.Helper()
	if _, ok := wire["type"]; !ok {
		t.Fatalf("wire packet missing type for lane=%q: %#v", lane, wire)
	}
	if _, ok := wire["lane"]; !ok {
		t.Fatalf("wire packet missing lane for lane=%q: %#v", lane, wire)
	}

	if hasOnlyKeys(wire, []string{"ships", "asteroids", "bullets", "pickups"}) {
		t.Fatalf("world payload encoded without envelope for lane=%q: %#v", lane, wire)
	}
	if hasOnlyKeys(wire, []string{"receiver"}) {
		t.Fatalf("overlay payload encoded without envelope for lane=%q: %#v", lane, wire)
	}
	if hasOnlyKeys(wire, []string{"players", "player_lifecycle", "total_asteroids"}) {
		t.Fatalf("session payload encoded without envelope for lane=%q: %#v", lane, wire)
	}
}

func hasOnlyKeys(wire map[string]any, keys []string) bool {
	if len(wire) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, ok := wire[key]; !ok {
			return false
		}
	}
	return true
}

func TestIncludedRealtimeLaneCandidatesReturnsOnlyIncludedCandidates(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull},
		{Lane: LaneOverlay, Kind: RealtimeLaneCandidateKindFull},
		{Lane: LaneSession, Kind: RealtimeLaneCandidateKindEventBatch},
	}
	included := []ScheduleRecord{
		{CandidateIndex: 0},
		{CandidateIndex: 2},
	}

	selected := IncludedRealtimeLaneCandidates(candidates, included)
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected candidates, got %d", len(selected))
	}
	if selected[0].Lane != LaneWorld || selected[1].Lane != LaneSession {
		t.Fatalf("selected candidates = %#v, want world then session", selected)
	}
}

func TestIncludedRealtimeLaneCandidatesPreservesIncludedOrder(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull},
		{Lane: LaneOverlay, Kind: RealtimeLaneCandidateKindFull},
		{Lane: LaneSession, Kind: RealtimeLaneCandidateKindEventBatch},
	}
	included := []ScheduleRecord{
		{CandidateIndex: 2},
		{CandidateIndex: 0},
	}

	selected := IncludedRealtimeLaneCandidates(candidates, included)
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected candidates, got %d", len(selected))
	}
	if selected[0].Lane != LaneSession || selected[1].Lane != LaneWorld {
		t.Fatalf("selected candidates = %#v, want session then world", selected)
	}
}

func TestIncludedRealtimeLaneCandidatesDeduplicatesRepeatedCandidateIndexes(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull},
		{Lane: LaneOverlay, Kind: RealtimeLaneCandidateKindFull},
	}
	included := []ScheduleRecord{
		{CandidateIndex: 1},
		{CandidateIndex: 1},
		{CandidateIndex: 0},
	}

	selected := IncludedRealtimeLaneCandidates(candidates, included)
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected candidates, got %d", len(selected))
	}
	if selected[0].Lane != LaneOverlay || selected[1].Lane != LaneWorld {
		t.Fatalf("selected candidates = %#v, want overlay then world", selected)
	}
}

func TestIncludedRealtimeLaneCandidatesSkipsInvalidIndexes(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull},
	}
	included := []ScheduleRecord{
		{CandidateIndex: -1},
		{CandidateIndex: 1},
		{CandidateIndex: 0},
	}

	selected := IncludedRealtimeLaneCandidates(candidates, included)
	if len(selected) != 1 {
		t.Fatalf("expected 1 selected candidate, got %d", len(selected))
	}
	if selected[0].Lane != LaneWorld {
		t.Fatalf("selected candidates = %#v, want world", selected)
	}
}
func TestEncodeLanePacketCompactsActiveWorldDeltaWireJSON(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: WorldDeltaPacket{
			Type: PacketTypeWorldDelta,
			Metadata: Metadata{
				Lane:         LaneWorld,
				Sequence:     9,
				BaselineID:   "baseline-9",
				SnapshotID:   "snapshot-9",
				SnapshotKind: SnapshotKind("delta"),
			},
			Ships: FieldRecordDelta[WorldShipRecord]{
				Updates: []map[string]any{{
					"id":        "ship-1",
					"x":         6,
					"y":         7,
					"rotation":  8,
					"thrusting": true,
				}},
			},
		},
	}

	encoded, recordedBytes := encodeLanePacket(candidate)
	if recordedBytes == 0 {
		t.Fatal("expected encoded bytes for active world delta packet")
	}
	if len(encoded) == 0 {
		t.Fatal("expected non-empty encoded packet")
	}

	wire := mustDecodeWirePacket(t, encoded)
	assertStringValue(t, wire, "t", "wd")
	assertStringValue(t, wire, "l", "w")
	assertContainsKey(t, wire, "q")
	assertContainsKey(t, wire, "su")
	assertNotContainsKey(t, wire, "server_sent_msec")
	assertNotContainsKey(t, wire, "snapshot_kind")
	assertNotContainsKey(t, wire, "ship_updates")
}
