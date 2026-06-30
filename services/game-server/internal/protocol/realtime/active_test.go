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
		Candidates: []RealtimeLaneCandidate{{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull}},
		SendPlan: SendPlan{Summary: summary},
		EncodedBytes: map[Lane]int{LaneWorld: 128},
		Mode: "active",
	}

	records := ActiveLaneMetricRecords(result)
	if len(records) != 1 {
		t.Fatalf("expected 1 metric record, got %d", len(records))
	}
	if records[0].Bytes != 128 {
		t.Fatalf("active metric bytes = %d, want 128", records[0].Bytes)
	}
}

func TestBuildActiveRealtimeResultEncodesOnlyEnvelopePackets(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Lives: 3,
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

	for _, candidate := range result.Candidates {
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
