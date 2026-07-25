package realtime

import (
	"bytes"
	"fmt"
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestActiveRealtimeSeparatesShipMovementAndLifecycleFromWorld(t *testing.T) {
	previous := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing", X: 10, Y: 20, Rotation: 0, Health: 3},
			"player-3": {ID: "player-3", ShipType: "v_wing", X: 30, Y: 40, Rotation: 0, Health: 3},
		},
	}
	current := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing", X: 40, Y: 80, Rotation: 1.5, Health: 3, Thrusting: true},
			"player-2": {ID: "player-2", ShipType: "v_wing", X: 100, Y: 200, Rotation: 0, Health: 3},
		},
	}

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline-1", SnapshotID: "world-baseline-1", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, mustWorldWireFull(t, previous, 1))

	result, err := BuildActiveRealtimeResult(current, state)
	if err != nil {
		t.Fatalf("BuildActiveRealtimeResult returned error: %v", err)
	}

	shipHot := requireCandidateByLane(t, result.SelectedCandidates, LaneShips)
	shipPacket, ok := shipHot.Payload.(ShipWireDeltaPacket)
	if !ok {
		t.Fatalf("ships payload type = %T", shipHot.Payload)
	}
	if len(shipPacket.ShipUpdates) != 1 || shipPacket.ShipUpdates[0]["id"] != "player-1" {
		t.Fatalf("ship hot updates = %#v, want player-1 movement", shipPacket.ShipUpdates)
	}

	lifecycle := requireCandidateByLane(t, result.SelectedCandidates, LaneShipsLifecycle)
	lifecyclePacket, ok := lifecycle.Payload.(ShipWireDeltaPacket)
	if !ok {
		t.Fatalf("ships.lifecycle payload type = %T", lifecycle.Payload)
	}
	if len(lifecyclePacket.ShipCreates) != 1 || lifecyclePacket.ShipCreates[0].ID != "player-2" {
		t.Fatalf("ship lifecycle creates = %#v, want player-2", lifecyclePacket.ShipCreates)
	}
	if len(lifecyclePacket.ShipDeletes) != 1 || lifecyclePacket.ShipDeletes[0] != "player-3" {
		t.Fatalf("ship lifecycle deletes = %#v, want player-3", lifecyclePacket.ShipDeletes)
	}

	world := requireCandidateByLane(t, result.SelectedCandidates, LaneWorld)
	worldPacket, ok := world.Payload.(WorldWireDeltaPacket)
	if !ok {
		t.Fatalf("world payload type = %T", world.Payload)
	}
	if len(worldPacket.Ships.Updates) != 0 || len(worldPacket.Ships.Creates) != 0 || len(worldPacket.Ships.Deletes) != 0 {
		t.Fatalf("world retained ship state: %#v", worldPacket.Ships)
	}

	assertLaneSchedulePolicy(t, result.PlannedRecords, LaneShips, DeliveryClassHotSupersedable, PriorityHigh)
	assertLaneSchedulePolicy(t, result.PlannedRecords, LaneShipsLifecycle, DeliveryClassRequired, PriorityCritical)
}

func TestShipPacketFamiliesUseDedicatedCompactTypes(t *testing.T) {
	metadata := Metadata{MatchID: "match-1", Lane: LaneShips, Sequence: 2, BaselineID: "world-baseline-1", SnapshotID: "ships-snapshot-2", SnapshotKind: SnapshotKind("delta"), IsFinalChunk: true}
	hot := mustRealtimeLaneCandidate(ShipWireDeltaPacket{
		Type:        PacketFamilyShipDelta,
		Metadata:    metadata,
		ShipUpdates: []map[string]any{{"id": "player-1", "x": int64(10), "y": int64(20), "rotation": int64(30), "thrusting": true}},
	}, nil)
	hotEncoded, _, err := encodeLanePacket(hot)
	if err != nil {
		t.Fatalf("encode ship hot packet: %v", err)
	}
	if !bytes.Contains(hotEncoded, []byte(`"t":"spd"`)) || !bytes.Contains(hotEncoded, []byte(`"su"`)) {
		t.Fatalf("ship hot compact packet = %s", hotEncoded)
	}

	metadata.Lane = LaneShipsLifecycle
	metadata.Sequence = 1
	metadata.SnapshotID = "ships-lifecycle-snapshot-1"
	lifecycle := mustRealtimeLaneCandidate(ShipWireDeltaPacket{
		Type:        PacketFamilyShipsLifecycle,
		Metadata:    metadata,
		ShipCreates: []WorldShipWireRecord{{ID: "player-2", ShipType: "v_wing", X: 10, Y: 20, Health: 3}},
	}, nil)
	lifecycleEncoded, _, err := encodeLanePacket(lifecycle)
	if err != nil {
		t.Fatalf("encode ship lifecycle packet: %v", err)
	}
	if !bytes.Contains(lifecycleEncoded, []byte(`"t":"spl"`)) || !bytes.Contains(lifecycleEncoded, []byte(`"sc"`)) {
		t.Fatalf("ship lifecycle compact packet = %s", lifecycleEncoded)
	}
}

func TestShipHotLaneChunksLargeBotCohortBelowHardCap(t *testing.T) {
	updates := make([]map[string]any, 0, 160)
	for index := 1; index <= 160; index++ {
		updates = append(updates, map[string]any{
			"id":        fmt.Sprintf("player-%d", index),
			"x":         int64(index * 10),
			"y":         int64(index * 20),
			"rotation":  int64(index * 3),
			"thrusting": index%2 == 0,
		})
	}
	candidate := mustRealtimeLaneCandidate(ShipWireDeltaPacket{
		Type:        PacketFamilyShipDelta,
		Metadata:    Metadata{MatchID: "match-1", Lane: LaneShips, Sequence: 5, BaselineID: "world-baseline-1", SnapshotID: "ships-snapshot-5", SnapshotKind: SnapshotKind("delta"), IsFinalChunk: true},
		ShipUpdates: updates,
	}, nil)

	chunks, err := ExpandRealtimeCandidateChunks([]RealtimeLaneCandidate{candidate})
	if err != nil {
		t.Fatalf("ExpandRealtimeCandidateChunks returned error: %v", err)
	}
	if len(chunks) <= 1 {
		t.Fatalf("ship chunks = %d, want more than one", len(chunks))
	}
	for index, chunk := range chunks {
		encoded, size, err := encodeLanePacket(chunk)
		if err != nil {
			t.Fatalf("encode ship chunk %d: %v", index, err)
		}
		if size > HardCapBytes {
			t.Fatalf("ship chunk %d size = %d, hard cap = %d, packet=%s", index, size, HardCapBytes, encoded)
		}
	}
}

func requireCandidateByLane(t *testing.T, candidates []RealtimeLaneCandidate, lane Lane) RealtimeLaneCandidate {
	t.Helper()
	for _, candidate := range candidates {
		if candidate.Lane() == lane {
			return candidate
		}
	}
	lanes := make([]Lane, 0, len(candidates))
	for _, candidate := range candidates {
		lanes = append(lanes, candidate.Lane())
	}
	t.Fatalf("missing candidate for lane %q; candidates=%v", lane, lanes)
	return RealtimeLaneCandidate{}
}

func assertLaneSchedulePolicy(t *testing.T, records []ScheduleRecord, lane Lane, delivery DeliveryClass, priority Priority) {
	t.Helper()
	for _, record := range records {
		if record.Lane != lane {
			continue
		}
		if record.DeliveryClass != delivery || record.Priority != priority {
			t.Fatalf("lane %q policy = delivery %q priority %q, want %q %q", lane, record.DeliveryClass, record.Priority, delivery, priority)
		}
		return
	}
	t.Fatalf("missing schedule record for lane %q", lane)
}
