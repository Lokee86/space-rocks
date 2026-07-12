package realtime

import (
	"fmt"
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
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
		SelectedCandidates: []RealtimeLaneCandidate{testCandidate(LaneWorld, RealtimeLaneCandidateKindFull)},
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
	candidate := testCandidate(LaneWorld, RealtimeLaneCandidateKindFull)
	candidate.Projection = "baseline-projection"

	projection, ok := CandidateProjection(candidate)
	if !ok {
		t.Fatal("expected full candidate projection to be returned")
	}
	if projection != "baseline-projection" {
		t.Fatalf("expected projection to match stored value, got %#v", projection)
	}
}

func TestCandidateProjectionReturnsFalseForNilProjection(t *testing.T) {
	candidate := testCandidate(LaneOverlay, RealtimeLaneCandidateKindFull)

	projection, ok := CandidateProjection(candidate)
	if ok || projection != nil {
		t.Fatalf("expected nil projection to be ignored, got %#v, %t", projection, ok)
	}
}

func TestCandidateProjectionReturnsFalseForEventBatchCandidate(t *testing.T) {
	candidate := testCandidate(LaneEvent, RealtimeLaneCandidateKindEventBatch)
	candidate.Projection = "should-be-ignored"

	projection, ok := CandidateProjection(candidate)
	if ok || projection != nil {
		t.Fatalf("expected event-batch candidate to have no projection, got %#v, %t", projection, ok)
	}
}

func decodedPacketType(wire map[string]any) string {
	if got, ok := wire["type"].(string); ok && got != "" {
		return got
	}
	compact, _ := wire["t"].(string)
	switch compact {
	case "wf":
		return "world_full"
	case "wd":
		return "world_delta"
	case "of":
		return "overlay_full"
	case "od":
		return "overlay_delta"
	case "sf":
		return "session_full"
	case "sd":
		return "session_delta"
	default:
		return ""
	}
}

func decodedPacketLane(wire map[string]any) string {
	if got, ok := wire["lane"].(string); ok && got != "" {
		return got
	}
	compact, _ := wire["l"].(string)
	switch compact {
	case "w":
		return "world"
	case "o":
		return "overlay"
	case "s":
		return "session"
	default:
		switch decodedPacketType(wire) {
		case "world_full", "world_delta":
			return "world"
		case "overlay_full", "overlay_delta":
			return "overlay"
		case "session_full", "session_delta":
			return "session"
		default:
			return ""
		}
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

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.UpdateLane(LaneSession, Metadata{Lane: LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)

	result := mustBuildActiveRealtimeResult(t, snapshot, state)
	if len(result.Candidates) == 0 {
		t.Fatal("expected active realtime result to emit candidates")
	}
	if len(result.SelectedCandidates) != len(result.Candidates) {
		t.Fatalf("expected selected candidates to match candidates in this baseline case, got %d selected and %d candidates", len(result.SelectedCandidates), len(result.Candidates))
	}

	for _, candidate := range result.SelectedCandidates {
		encodedPacket, ok := result.EncodedPackets[candidate.Lane()]
		if !ok || len(encodedPacket) == 0 {
			t.Fatalf("expected encoded packet for lane=%q kind=%q", candidate.Lane(), candidate.Kind())
		}
		wire := mustDecodeWirePacket(t, encodedPacket)
		if candidate.Lane() == LaneEvent {
			assertStringValue(t, wire, "t", "eb")
			assertStringValue(t, wire, "type", PacketFamilyEventBatch)
			assertContainsKey(t, wire, "bid")
			assertContainsKey(t, wire, "ev")
			assertNotContainsKey(t, wire, "lane")
			assertNotContainsKey(t, wire, "baseline_id")
			assertNotContainsKey(t, wire, "snapshot_id")
			assertNotContainsKey(t, wire, "snapshot_kind")
			assertNotContainsKey(t, wire, "chunk_index")
			assertNotContainsKey(t, wire, "chunk_count")
			assertNotContainsKey(t, wire, "is_final_chunk")
			continue
		}
		if decodedPacketType(wire) == "" {
			t.Fatalf("expected non-empty top-level type for lane=%q kind=%q, got %#v", candidate.Lane(), candidate.Kind(), wire)
		}
		if decodedPacketLane(wire) == "" {
			t.Fatalf("expected non-empty top-level lane for lane=%q kind=%q, got %#v", candidate.Lane(), candidate.Kind(), wire)
		}
		if candidate.Lane() == LaneWorld {
			asteroids, ok := wire["asteroids"].([]any)
			if !ok || len(asteroids) == 0 {
				t.Fatalf("expected compact world packet to include asteroids, got %#v", wire["asteroids"])
			}
			tuple, ok := asteroids[0].([]any)
			if !ok || len(tuple) == 0 {
				t.Fatalf("expected compact asteroid tuple, got %#v", asteroids[0])
			}
			if !hasNumericCompactAsteroidTupleID(tuple[0]) {
				t.Fatalf("expected compact asteroid tuple id to be numeric suffix 1, not a string, got %#v", tuple[0])
			}
		}
		assertNotNakedLanePayload(t, candidate.Lane(), wire)
	}
}

func TestBuildActiveRealtimeResultEncodesMultipleAsteroidLanePackets(t *testing.T) {
	previousAsteroids := make(map[string]runtime.AsteroidState, 300)
	currentAsteroids := make(map[string]runtime.AsteroidState, 300)
	for i := 1; i <= 300; i++ {
		id := fmt.Sprintf("asteroid-%06d", i)
		previousAsteroids[id] = runtime.AsteroidState{ID: id, X: float64(i), Y: float64(i + 1), Size: 2, Health: 3, Scale: 1.25, Variant: 4}
		currentAsteroids[id] = runtime.AsteroidState{ID: id, X: float64(i + 20), Y: float64(i + 21), Size: 2, Health: 3, Scale: 1.25, Variant: 4}
	}

	previousSnapshot := game.GameplayPresentationSnapshot{
		SelfID:         "player-1",
		Lives:          3,
		ServerSentMsec: 2234,
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing", X: 1, Y: 2, Rotation: 3, Health: 4, Shields: 5},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 9, Lives: 3, RespawnCooldown: 1.25, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite", SecondaryWeaponID: "mine", SecondaryAmmoPolicy: "limited"},
		},
		PlayerLifecycle: map[string]string{"player-1": "active"},
		Asteroids:       previousAsteroids,
	}
	currentSnapshot := previousSnapshot
	currentSnapshot.ServerSentMsec = 2235
	currentSnapshot.Asteroids = currentAsteroids

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, previousSnapshot, 1))
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.StoreBaselineProjection(LaneOverlay, mustOverlayWireFull(t, previousSnapshot, "player-1", 1))
	state.UpdateLane(LaneSession, Metadata{Lane: LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.StoreBaselineProjection(LaneSession, mustSessionWireFull(t, previousSnapshot, 1))

	result := mustBuildActiveRealtimeResult(t, currentSnapshot, state)

	asteroidPackets := encodedPacketsForLane(result, LaneAsteroids)
	if len(asteroidPackets) <= 1 {
		t.Fatalf("expected multiple asteroid lane packets, got %d", len(asteroidPackets))
	}

	selectedAsteroidCandidates := 0
	for _, candidate := range result.SelectedCandidates {
		if candidate.Lane() == LaneAsteroids {
			selectedAsteroidCandidates++
		}
	}
	if selectedAsteroidCandidates != len(asteroidPackets) {
		t.Fatalf("selected asteroid candidates mismatch: candidates=%d selected=%d encoded=%d selected_asteroid_candidates=%d asteroid_packets=%d total_asteroid_updates=%d", len(result.Candidates), len(result.SelectedCandidates), len(result.EncodedLanePackets), selectedAsteroidCandidates, len(asteroidPackets), 0)
	}

	totalAsteroidUpdates := 0
	for index, encoded := range asteroidPackets {
		if encoded.Candidate.Lane() != LaneAsteroids {
			t.Fatalf("asteroid encoded packet %d lane = %q, want asteroids", index, encoded.Candidate.Lane())
		}
		if encoded.Candidate.Kind() != RealtimeLaneCandidateKindDelta {
			t.Fatalf("asteroid encoded packet %d kind = %q, want delta", index, encoded.Candidate.Kind())
		}
		packet, ok := encoded.Candidate.Payload.(AsteroidWireDeltaPacket)
		if !ok {
			t.Fatalf("asteroid encoded packet %d payload type = %T, want AsteroidWireDeltaPacket", index, encoded.Candidate.Payload)
		}
		totalAsteroidUpdates += len(packet.AsteroidUpdates)
		if len(packet.AsteroidUpdates) != 1 && encoded.EncodedBytes > HardCapBytes {
			t.Fatalf("asteroid encoded packet %d bytes = %d, want <= %d unless single-update chunk", index, encoded.EncodedBytes, HardCapBytes)
		}
	}
	if totalAsteroidUpdates != len(previousAsteroids) {
		t.Fatalf("asteroid updates across chunks = %d, want %d (candidates=%d selected=%d encoded=%d selected_asteroid_candidates=%d asteroid_packets=%d total_asteroid_updates=%d)", totalAsteroidUpdates, len(previousAsteroids), len(result.Candidates), len(result.SelectedCandidates), len(result.EncodedLanePackets), selectedAsteroidCandidates, len(asteroidPackets), totalAsteroidUpdates)
	}

	if result.TotalEncodedBytes <= 0 {
		t.Fatal("expected total encoded bytes to be positive")
	}

	asteroidMetricCount := 0
	for _, record := range result.MetricSummaries {
		if record.PacketFamily == PacketFamilyAsteroidDelta {
			asteroidMetricCount++
		}
	}
	if asteroidMetricCount <= 1 {
		t.Fatalf("expected more than one asteroid metric summary, got %d", asteroidMetricCount)
	}
}

func TestBuildActiveRealtimeResultEncodesMultipleBulletLanePackets(t *testing.T) {
	previousBullets := make(map[string]runtime.BulletState, 240)
	currentBullets := make(map[string]runtime.BulletState, 240)
	for i := 1; i <= 240; i++ {
		id := fmt.Sprintf("bullet-%06d", i)
		previousBullets[id] = runtime.BulletState{ID: id, OwnerID: "player-1", X: float64(i), Y: float64(i + 1), Rotation: float64(i + 2), WeaponID: "laser", ProjectileType: "bolt"}
		currentBullets[id] = runtime.BulletState{ID: id, OwnerID: "player-1", X: float64(i + 10), Y: float64(i + 11), Rotation: float64(i + 12), WeaponID: "laser", ProjectileType: "bolt"}
	}

	previousSnapshot := game.GameplayPresentationSnapshot{
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
		Bullets:         previousBullets,
	}
	currentSnapshot := previousSnapshot
	currentSnapshot.ServerSentMsec = 1235
	currentSnapshot.Bullets = currentBullets

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, previousSnapshot, 1))
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.StoreBaselineProjection(LaneOverlay, mustOverlayWireFull(t, previousSnapshot, "player-1", 1))
	state.UpdateLane(LaneSession, Metadata{Lane: LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.StoreBaselineProjection(LaneSession, mustSessionWireFull(t, previousSnapshot, 1))

	result := mustBuildActiveRealtimeResult(t, currentSnapshot, state)

	bulletPackets := encodedPacketsForLane(result, LaneBullets)
	if len(bulletPackets) <= 1 {
		t.Fatalf("expected multiple bullet lane packets, got %d", len(bulletPackets))
	}

	selectedBulletCandidates := 0
	for _, candidate := range result.SelectedCandidates {
		if candidate.Lane() == LaneBullets {
			selectedBulletCandidates++
		}
	}
	diagnostics := func(totalBulletUpdates int) string {
		return fmt.Sprintf("candidates=%d selected=%d encoded=%d selected_bullet_candidates=%d bullet_packets=%d total_bullet_updates=%d", len(result.Candidates), len(result.SelectedCandidates), len(result.EncodedLanePackets), selectedBulletCandidates, len(bulletPackets), totalBulletUpdates)
	}
	if selectedBulletCandidates != len(bulletPackets) {
		t.Fatalf("selected bullet candidates mismatch: %s", diagnostics(0))
	}

	totalBulletUpdates := 0
	for index, encoded := range bulletPackets {
		if encoded.Candidate.Lane() != LaneBullets {
			t.Fatalf("bullet encoded packet %d lane = %q, want bullets", index, encoded.Candidate.Lane())
		}
		if encoded.Candidate.Kind() != RealtimeLaneCandidateKindDelta {
			t.Fatalf("bullet encoded packet %d kind = %q, want delta", index, encoded.Candidate.Kind())
		}
		packet, ok := encoded.Candidate.Payload.(BulletWireDeltaPacket)
		if !ok {
			t.Fatalf("bullet encoded packet %d payload type = %T, want BulletWireDeltaPacket", index, encoded.Candidate.Payload)
		}
		totalBulletUpdates += len(packet.BulletUpdates)
		if len(packet.BulletUpdates) != 1 && encoded.EncodedBytes > HardCapBytes {
			t.Fatalf("bullet encoded packet %d bytes = %d, want <= %d unless single-update chunk", index, encoded.EncodedBytes, HardCapBytes)
		}
	}
	if totalBulletUpdates != len(previousBullets) {
		t.Fatalf("bullet updates across chunks = %d, want %d (%s)", totalBulletUpdates, len(previousBullets), diagnostics(totalBulletUpdates))
	}

	if result.TotalEncodedBytes <= 0 {
		t.Fatal("expected total encoded bytes to be positive")
	}

	bulletMetricCount := 0
	for _, record := range result.MetricSummaries {
		if record.PacketFamily == PacketFamilyBulletDelta {
			bulletMetricCount++
		}
	}
	if bulletMetricCount <= 1 {
		t.Fatalf("expected more than one bullet metric summary, got %d", bulletMetricCount)
	}
}

func TestBuildActiveRealtimeResultSelectsFullPacketsWithoutStoredBaselines(t *testing.T) {

	snapshot := tinyActiveBoundarySnapshot()
	result := mustBuildActiveRealtimeResult(t, snapshot, NewRealtimeSessionState("player-1", "match-1"))

	assertSelectedCandidate(t, result, LaneWorld, RealtimeLaneCandidateKindFull)
	assertSelectedCandidate(t, result, LaneOverlay, RealtimeLaneCandidateKindFull)
	assertSelectedCandidate(t, result, LaneSession, RealtimeLaneCandidateKindFull)

	assertEncodedPacketTypeAndLane(t, result, LaneWorld, PacketFamilyWorldFull, string(LaneWorld))
	assertEncodedPacketTypeAndLane(t, result, LaneOverlay, PacketFamilyOverlayFull, string(LaneOverlay))
	assertEncodedPacketTypeAndLane(t, result, LaneSession, PacketFamilySessionFull, string(LaneSession))
}

func TestBuildActiveRealtimeResultEmitsNoWorldOverlayOrSessionPacketsWhenStoredBaselinesMatch(t *testing.T) {
	snapshot := tinyActiveBoundarySnapshot()
	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, snapshot, 1))
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.StoreBaselineProjection(LaneOverlay, mustOverlayWireFull(t, snapshot, "player-1", 1))
	state.UpdateLane(LaneSession, Metadata{Lane: LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.StoreBaselineProjection(LaneSession, mustSessionWireFull(t, snapshot, 1))

	result := mustBuildActiveRealtimeResult(t, snapshot, state)
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

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 2, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, game.GameplayPresentationSnapshot{SelfID: "player-1", Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing", X: 1, Y: 1, Rotation: 0, Health: 5, Shields: 0}}}, 1))

	result := mustBuildActiveRealtimeResult(t, snapshot, state)
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
		if candidate.Lane() == lane && candidate.Kind() == kind {
			return
		}
	}
	t.Fatalf("expected selected candidate lane=%q kind=%q, got %#v", lane, kind, result.SelectedCandidates)
}

func assertNoSelectedCandidate(t *testing.T, result ActiveRealtimeResult, lane Lane) {
	t.Helper()
	for _, candidate := range result.SelectedCandidates {
		if candidate.Lane() == lane {
			t.Fatalf("expected no selected candidate for lane=%q, got %#v", lane, result.SelectedCandidates)
		}
	}
	if _, ok := result.EncodedPackets[lane]; ok {
		t.Fatalf("expected no encoded packet for lane=%q, got %#v", lane, result.EncodedPackets[lane])
	}
}

func encodedPacketsForLane(result ActiveRealtimeResult, lane Lane) []EncodedRealtimeLanePacket {
	packets := make([]EncodedRealtimeLanePacket, 0, len(result.EncodedLanePackets))
	for _, packet := range result.EncodedLanePackets {
		if packet.Candidate.Lane() == lane {
			packets = append(packets, packet)
		}
	}
	return packets
}

func assertEncodedPacketTypeAndLane(t *testing.T, result ActiveRealtimeResult, lane Lane, wantType string, wantLane string) {
	t.Helper()
	encoded, ok := result.EncodedPackets[lane]
	if !ok || len(encoded) == 0 {
		t.Fatalf("expected encoded packet for lane=%q", lane)
	}
	wire := mustDecodeWirePacket(t, encoded)
	if got := decodedPacketType(wire); got != wantType {
		t.Fatalf("expected type=%q for lane=%q, got %#v", wantType, lane, wire)
	}
	if got := decodedPacketLane(wire); got != wantLane {
		t.Fatalf("expected lane=%q for lane=%q, got %#v", wantLane, lane, wire)
	}
}

func TestIncludedRealtimeLaneCandidatesSkipsDeferredRecordsInOrder(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		testCandidate(LaneWorld, RealtimeLaneCandidateKindFull),
		testCandidate(LaneOverlay, RealtimeLaneCandidateKindFull),
		testCandidate(LaneSession, RealtimeLaneCandidateKindDelta),
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
	if selected[0].Lane() != LaneSession || selected[1].Lane() != LaneWorld {
		t.Fatalf("selected candidates = %#v, want session then world", selected)
	}
}

func TestBuildActiveRealtimeResultUsesSelectedCandidatesOnly(t *testing.T) {
	result := ActiveRealtimeResult{
		SelectedCandidates: []RealtimeLaneCandidate{
			testCandidate(LaneOverlay, RealtimeLaneCandidateKindFull),
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
		if _, ok := wire["t"]; !ok {
			t.Fatalf("wire packet missing type for lane=%q: %#v", lane, wire)
		}
	}
	if decodedPacketType(wire) == "" {
		t.Fatalf("wire packet missing decodable envelope type for lane=%q: %#v", lane, wire)
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

func hasNumericCompactAsteroidTupleID(value any) bool {
	switch got := value.(type) {
	case int:
		return got == 1
	case int64:
		return got == 1
	case float64:
		return got == 1
	default:
		return false
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
		testCandidate(LaneWorld, RealtimeLaneCandidateKindFull),
		testCandidate(LaneOverlay, RealtimeLaneCandidateKindFull),
		testCandidate(LaneSession, RealtimeLaneCandidateKindDelta),
	}
	included := []ScheduleRecord{
		{CandidateIndex: 0},
		{CandidateIndex: 2},
	}

	selected := IncludedRealtimeLaneCandidates(candidates, included)
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected candidates, got %d", len(selected))
	}
	if selected[0].Lane() != LaneWorld || selected[1].Lane() != LaneSession {
		t.Fatalf("selected candidates = %#v, want world then session", selected)
	}
}

func TestIncludedRealtimeLaneCandidatesPreservesIncludedOrder(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		testCandidate(LaneWorld, RealtimeLaneCandidateKindFull),
		testCandidate(LaneOverlay, RealtimeLaneCandidateKindFull),
		testCandidate(LaneSession, RealtimeLaneCandidateKindDelta),
	}
	included := []ScheduleRecord{
		{CandidateIndex: 2},
		{CandidateIndex: 0},
	}

	selected := IncludedRealtimeLaneCandidates(candidates, included)
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected candidates, got %d", len(selected))
	}
	if selected[0].Lane() != LaneSession || selected[1].Lane() != LaneWorld {
		t.Fatalf("selected candidates = %#v, want session then world", selected)
	}
}

func TestIncludedRealtimeLaneCandidatesDeduplicatesRepeatedCandidateIndexes(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		testCandidate(LaneWorld, RealtimeLaneCandidateKindFull),
		testCandidate(LaneOverlay, RealtimeLaneCandidateKindFull),
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
	if selected[0].Lane() != LaneOverlay || selected[1].Lane() != LaneWorld {
		t.Fatalf("selected candidates = %#v, want overlay then world", selected)
	}
}

func TestBuildActiveRealtimeResultRecoversInvalidatedWorldBaseline(t *testing.T) {
	snapshot := tinyActiveBoundarySnapshot()
	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 4, BaselineID: "world-baseline-4", SnapshotID: "world-snapshot-4", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, snapshot, 4))
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 3, BaselineID: "overlay-baseline-3", SnapshotID: "overlay-snapshot-3", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.StoreBaselineProjection(LaneOverlay, mustOverlayWireFull(t, snapshot, "player-1", 3))
	if !state.RequireFullBaseline(LaneWorld) {
		t.Fatal("expected invalidation")
	}
	result := mustBuildActiveRealtimeResult(t, snapshot, state)
	var world RealtimeLaneCandidate
	found := false
	for _, candidate := range result.SelectedCandidates {
		if candidate.Lane() == LaneWorld {
			world, found = candidate, true
			break
		}
	}
	if !found || world.Kind() != RealtimeLaneCandidateKindFull {
		t.Fatalf("expected world full candidate, got %#v", result.SelectedCandidates)
	}
	metadata, ok := world.Metadata()
	if !ok || metadata.Sequence != 5 || metadata.BaselineID != FullBaselineID(LaneWorld, 5) || metadata.BaselineID == "world-baseline-4" {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	if !state.LaneBaselineReady(LaneOverlay) {
		t.Fatal("expected overlay ready")
	}
	for _, candidate := range result.SelectedCandidates {
		if candidate.Lane() == LaneOverlay && candidate.Kind() == RealtimeLaneCandidateKindFull {
			t.Fatal("overlay forced full")
		}
	}
}

func TestIncludedRealtimeLaneCandidatesSkipsInvalidIndexes(t *testing.T) {
	candidates := []RealtimeLaneCandidate{
		testCandidate(LaneWorld, RealtimeLaneCandidateKindFull),
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
	if selected[0].Lane() != LaneWorld {
		t.Fatalf("selected candidates = %#v, want world", selected)
	}
}
func TestEncodeLanePacketCompactsActiveWorldDeltaWireJSON(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(WorldDeltaPacket{
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
	}, nil)

	encoded, recordedBytes := mustEncodeLanePacket(t, candidate)
	if recordedBytes == 0 {
		t.Fatal("expected encoded bytes for active world delta packet")
	}
	if len(encoded) == 0 {
		t.Fatal("expected non-empty encoded packet")
	}

	wire := mustDecodeWirePacket(t, encoded)
	assertStringValue(t, wire, "t", "wd")
	assertContainsKey(t, wire, "q")
	assertContainsKey(t, wire, "su")
	assertNotContainsKey(t, wire, "l")
	assertNotContainsKey(t, wire, "server_sent_msec")
	assertNotContainsKey(t, wire, "snapshot_kind")
	assertNotContainsKey(t, wire, "k")
	assertNotContainsKey(t, wire, "sid")
	assertNotContainsKey(t, wire, "ship_updates")
}
