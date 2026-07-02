package realtime

import (
	"encoding/json"
	"strings"
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	runtime "github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

func TestActiveWirePacketEncodingUsesLowercaseWorldShape(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindFull,
		Full: WorldFullPacket{
			Type: PacketFamilyWorldFull,
			Metadata: Metadata{
				Lane:     LaneWorld,
				Sequence: 7,
			},
			Ships: []WorldShipRecord{
				{
					ID:         "ship-1",
					ShipType:   "v_wing",
					X:          1,
					Y:          2,
					Rotation:   3,
					Health:     4,
					Shields:    5,
					Thrusting:  true,
					TargetKind: "player",
					TargetID:   "player-1",
				},
			},
			Bullets: []WorldBulletRecord{
				{
					ID:             "bullet-1",
					OwnerID:        "ship-1",
					X:              6,
					Y:              7,
					Rotation:       8,
					WeaponID:       "basic",
					ProjectileType: "laser",
				},
			},
			Asteroids: []WorldAsteroidRecord{
				{
					ID:      "asteroid-1",
					X:       9,
					Y:       10,
					Size:    2,
					Health:  11,
					Scale:   1.5,
					Variant: 3,
				},
			},
			Pickups: []WorldPickupRecord{
				{
					ID:              "pickup-1",
					Type:            "shield",
					PickupClass:     "armor",
					X:               12,
					Y:               13,
					Health:          1,
					AgeSeconds:      4.5,
					LifespanSeconds: 9.5,
				},
			},
		},
	}

	encoded := mustEncodeWirePacket(t, candidate)
	wire := mustDecodeWirePacket(t, encoded)

	assertStringValue(t, wire, "type", PacketFamilyWorldFull)
	assertContainsKey(t, wire, "ships")
	assertNotContainsKey(t, wire, "Type")
	assertNotContainsKey(t, wire, "Metadata")
	assertNotContainsKey(t, wire, "Ships")

	ships := mustSliceValue(t, wire, "ships")
	ship := mustMapValue(t, ships[0])
	assertStringValue(t, ship, "id", "ship-1")
	assertStringValue(t, ship, "ship_type", "v_wing")
	assertNotContainsKey(t, ship, "ShipType")
	assertNotContainsKey(t, ship, "ID")

	asteroids := mustSliceValue(t, wire, "asteroids")
	asteroid := mustMapValue(t, asteroids[0])
	assertFloatValue(t, asteroid, "scale", 1.5)
	assertIntValue(t, asteroid, "variant", 3)
}


func TestWorldQuantizationReachesEncodedWireJSON(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"ship-1": {
				ID:       "ship-1",
				ShipType: "v_wing",
				X:        123.456789,
				Y:        987.654321,
				Rotation: 3.1415926535,
			},
		},
		Bullets: map[string]runtime.BulletState{
			"bullet-1": {
				ID:       "bullet-1",
				OwnerID:  "ship-1",
				X:        11.111111,
				Y:        22.222222,
				Rotation: 1.23456789,
			},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-1": {
				ID:    "asteroid-1",
				X:     333.333333,
				Y:     444.444444,
				Scale: 1.23456789,
			},
		},
		Pickups: map[string]runtime.PickupState{
			"pickup-1": {
				ID:              "pickup-1",
				Type:            "shield",
				X:               55.555555,
				Y:               66.666666,
				AgeSeconds:      7.891234,
				LifespanSeconds: 12.345678,
			},
		},
	}

	full := BuildWorldFullPacket(snapshot, 1)
	wire, err := quantizeWorldFullPacket(full)
	if err != nil {
		t.Fatalf("quantize world full packet: %v", err)
	}

	encoded, err := packetcodec.Encode(WireLanePacket(RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull, Full: wire}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	encodedString := string(encoded)
	for _, fragment := range []string{"123.456789", "987.654321", "3.1415926535", "1.23456789", "7.891234", "12.345678"} {
		if strings.Contains(encodedString, fragment) {
			t.Fatalf("expected encoded JSON to not contain raw fragment %q, got %s", fragment, encodedString)
		}
	}

	decoded := mustDecodeWirePacket(t, encoded)
	ships := mustSliceValue(t, decoded, "ships")
	ship := mustMapValue(t, ships[0])
	assertJSONIntValue(t, ship, "x", 1235)
	assertJSONIntValue(t, ship, "y", 9877)
	assertJSONIntValue(t, ship, "rotation", 3142)

	bullets := mustSliceValue(t, decoded, "bullets")
	bullet := mustMapValue(t, bullets[0])
	assertJSONIntValue(t, bullet, "x", 111)
	assertJSONIntValue(t, bullet, "y", 222)
	assertJSONIntValue(t, bullet, "rotation", 1235)

	asteroids := mustSliceValue(t, decoded, "asteroids")
	asteroid := mustMapValue(t, asteroids[0])
	assertJSONIntValue(t, asteroid, "x", 3333)
	assertJSONIntValue(t, asteroid, "y", 4444)
	assertJSONIntValue(t, asteroid, "scale", 1235)

	pickups := mustSliceValue(t, decoded, "pickups")
	pickup := mustMapValue(t, pickups[0])
	assertJSONIntValue(t, pickup, "x", 556)
	assertJSONIntValue(t, pickup, "y", 667)
	assertJSONIntValue(t, pickup, "age_seconds", 7891)
	assertJSONIntValue(t, pickup, "lifespan_seconds", 12346)
}

func TestWireWorldWireDeltaPacketOmitsEmptySectionsAndKeepsShipUpdates(t *testing.T) {
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: WorldWireDeltaPacket{
			Type: PacketTypeWorldDelta,
			Metadata: Metadata{Lane: LaneWorld, Sequence: 9, SnapshotKind: SnapshotKind("delta")},
			Ships: FieldRecordDelta[WorldShipWireRecord]{
				Updates: []map[string]any{
					{
						"id":        "ship-1",
						"x":         int64(10),
						"y":         int64(20),
						"rotation":  int64(30),
						"thrusting": true,
					},
				},
			},
		},
	}))

	updates := mustSliceValue(t, wire, "ship_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one ship update, got %#v", updates)
	}

	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "id", "ship-1")
	assertJSONIntValue(t, update, "x", 10)
	assertJSONIntValue(t, update, "y", 20)
	assertJSONIntValue(t, update, "rotation", 30)
	if thrusting, ok := update["thrusting"].(bool); !ok || !thrusting {
		t.Fatalf("expected thrusting to be true, got %#v", update["thrusting"])
	}

	for _, key := range []string{"ship_creates", "ship_deletes", "bullet_creates", "bullet_updates", "bullet_deletes", "asteroid_creates", "asteroid_updates", "asteroid_deletes", "pickup_creates", "pickup_updates", "pickup_deletes"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireWorldWireDeltaPacketOmitsEmptySectionsAndKeepsBulletDeletes(t *testing.T) {
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: WorldWireDeltaPacket{
			Type: PacketTypeWorldDelta,
			Metadata: Metadata{Lane: LaneWorld, Sequence: 10, SnapshotKind: SnapshotKind("delta")},
			Bullets: FieldRecordDelta[WorldBulletWireRecord]{
				Deletes: []string{"bullet-1"},
			},
		},
	}))

	deletes := mustSliceValue(t, wire, "bullet_deletes")
	if len(deletes) != 1 {
		t.Fatalf("expected one bullet delete, got %#v", deletes)
	}
	if got, ok := deletes[0].(string); !ok || got != "bullet-1" {
		t.Fatalf("expected bullet delete %q, got %#v", "bullet-1", deletes[0])
	}

	for _, key := range []string{"ship_creates", "ship_updates", "ship_deletes", "bullet_creates", "bullet_updates", "asteroid_creates", "asteroid_updates", "asteroid_deletes", "pickup_creates", "pickup_updates", "pickup_deletes"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireWorldWireDeltaPacketPreservesFalseAndZeroFieldsInShipUpdates(t *testing.T) {
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: WorldWireDeltaPacket{
			Type: PacketTypeWorldDelta,
			Metadata: Metadata{Lane: LaneWorld, Sequence: 11, SnapshotKind: SnapshotKind("delta")},
			Ships: FieldRecordDelta[WorldShipWireRecord]{
				Updates: []map[string]any{
					{
						"id":        "ship-1",
						"x":         int64(0),
						"y":         int64(0),
						"rotation":  int64(0),
						"thrusting": false,
					},
				},
			},
		},
	}))

	update := mustMapValue(t, mustSliceValue(t, wire, "ship_updates")[0])
	assertJSONIntValue(t, update, "x", 0)
	assertJSONIntValue(t, update, "y", 0)
	assertJSONIntValue(t, update, "rotation", 0)
	if thrusting, ok := update["thrusting"].(bool); !ok || thrusting {
		t.Fatalf("expected thrusting to be false, got %#v", update["thrusting"])
	}
}

func TestWireWorldWireDeltaPacketSparseJsonOmitsEmptySections(t *testing.T) {
	encoded, err := packetcodec.Encode(WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: WorldWireDeltaPacket{
			Type: PacketTypeWorldDelta,
			Metadata: Metadata{Lane: LaneWorld, Sequence: 20, SnapshotKind: SnapshotKind("delta")},
			Ships: FieldRecordDelta[WorldShipWireRecord]{
				Updates: []map[string]any{{"id": "ship-1", "x": int64(1), "y": int64(2), "rotation": int64(3), "thrusting": true}},
			},
		},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	encodedString := string(encoded)
	for _, fragment := range []string{"ship_creates", "bullet_creates", "bullet_updates", "asteroid_updates", "pickup_deletes"} {
		if strings.Contains(encodedString, fragment) {
			t.Fatalf("expected encoded JSON to omit %q, got %s", fragment, encodedString)
		}
	}
	for _, fragment := range []string{"ship_updates", "ship-1"} {
		if !strings.Contains(encodedString, fragment) {
			t.Fatalf("expected encoded JSON to contain %q, got %s", fragment, encodedString)
		}
	}
}

func TestWireOverlayWireDeltaPacketSparseJsonOmitsEmptySections(t *testing.T) {
	encoded, err := packetcodec.Encode(WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneOverlay,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: OverlayWireLaneDelta{
			Metadata: Metadata{Lane: LaneOverlay, Sequence: 21, SnapshotKind: SnapshotKind("delta")},
		},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	encodedString := string(encoded)
	for _, fragment := range []string{"receiver_creates", "receiver_updates", "receiver_deletes"} {
		if strings.Contains(encodedString, fragment) {
			t.Fatalf("expected encoded JSON to omit %q, got %s", fragment, encodedString)
		}
	}
}

func TestWireSessionWireDeltaPacketSparseJsonOmitsEmptySections(t *testing.T) {
	encoded, err := packetcodec.Encode(WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneSession,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: SessionWireLaneDelta{
			Metadata: Metadata{Lane: LaneSession, Sequence: 22, SnapshotKind: SnapshotKind("delta")},
		},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	encodedString := string(encoded)
	for _, fragment := range []string{"player_session_updates", "player_lifecycle_updates", "total_asteroids"} {
		if strings.Contains(encodedString, fragment) {
			t.Fatalf("expected encoded JSON to omit %q, got %s", fragment, encodedString)
		}
	}
}

func TestWireWorldDeltaPacketJSONDoesNotContainNullForEmptyDelta(t *testing.T) {
	encoded, err := packetcodec.Encode(wireWorldDeltaPacket(WorldDeltaPacket{Type: PacketTypeWorldDelta}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if strings.Contains(string(encoded), "null") {
		t.Fatalf("expected empty world delta JSON not to contain null, got %s", string(encoded))
	}
}

func TestWireWorldDeltaPacketEncodesShipUpdatesAsPartialFieldPatch(t *testing.T) {
	encoded, err := packetcodec.Encode(wireWorldDeltaPacket(WorldDeltaPacket{
		Type: PacketTypeWorldDelta,
		Ships: FieldRecordDelta[WorldShipRecord]{
			Updates: []map[string]any{
				{
					"id":         "ship-1",
					"x":          6,
					"y":          7,
					"rotation":   8,
					"thrusting":  true,
				},
			},
		},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	wire := mustDecodeWirePacket(t, encoded)
	updates := mustSliceValue(t, wire, "ship_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one ship update, got %#v", updates)
	}

	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "id", "ship-1")
	assertFloatValue(t, update, "x", 6)
	assertFloatValue(t, update, "y", 7)
	assertFloatValue(t, update, "rotation", 8)
	assertNotContainsKey(t, update, "ship_type")
	assertNotContainsKey(t, update, "health")
	assertNotContainsKey(t, update, "shields")
	assertNotContainsKey(t, update, "target_kind")
	assertNotContainsKey(t, update, "target_id")
}

func TestWireWorldDeltaPacketEncodesBulletUpdatesAsPartialFieldPatch(t *testing.T) {
	encoded, err := packetcodec.Encode(wireWorldDeltaPacket(WorldDeltaPacket{
		Type: PacketTypeWorldDelta,
		Bullets: FieldRecordDelta[WorldBulletRecord]{
			Updates: []map[string]any{
				{
					"id":       "bullet-1",
					"x":        6,
					"y":        7,
					"rotation": 8,
				},
			},
		},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	wire := mustDecodeWirePacket(t, encoded)
	updates := mustSliceValue(t, wire, "bullet_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one bullet update, got %#v", updates)
	}

	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "id", "bullet-1")
	assertFloatValue(t, update, "x", 6)
	assertFloatValue(t, update, "y", 7)
	assertFloatValue(t, update, "rotation", 8)
	assertNotContainsKey(t, update, "owner_id")
	assertNotContainsKey(t, update, "weapon_id")
	assertNotContainsKey(t, update, "projectile_type")
}

func TestWireWorldDeltaPacketEncodesBulletUpdatesWithZeroRotation(t *testing.T) {
	encoded, err := packetcodec.Encode(wireWorldDeltaPacket(WorldDeltaPacket{
		Type: PacketTypeWorldDelta,
		Bullets: FieldRecordDelta[WorldBulletRecord]{
			Updates: []map[string]any{
				{
					"id":       "bullet-1",
					"x":        6,
					"y":        7,
					"rotation": 0,
				},
			},
		},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	wire := mustDecodeWirePacket(t, encoded)
	update := mustMapValue(t, mustSliceValue(t, wire, "bullet_updates")[0])
	assertFloatValue(t, update, "rotation", 0)
	assertNotContainsKey(t, update, "weapon_id")
	assertNotContainsKey(t, update, "projectile_type")
}

func TestWireWorldDeltaPacketEncodesAsteroidUpdatesAsPartialFieldPatch(t *testing.T) {
	encoded, err := packetcodec.Encode(wireWorldDeltaPacket(WorldDeltaPacket{
		Type: PacketTypeWorldDelta,
		Asteroids: FieldRecordDelta[WorldAsteroidRecord]{
			Updates: []map[string]any{
				{
					"id":     "asteroid-1",
					"x":      6,
					"y":      7,
					"size":   2,
					"health": 11,
				},
			},
		},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	wire := mustDecodeWirePacket(t, encoded)
	updates := mustSliceValue(t, wire, "asteroid_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one asteroid update, got %#v", updates)
	}

	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "id", "asteroid-1")
	assertFloatValue(t, update, "x", 6)
	assertFloatValue(t, update, "y", 7)
	assertNotContainsKey(t, update, "size")
	assertNotContainsKey(t, update, "health")
	assertNotContainsKey(t, update, "scale")
	assertNotContainsKey(t, update, "variant")
}

func TestWireWorldDeltaPacketEncodesPickupUpdatesAsPartialFieldPatch(t *testing.T) {
	encoded, err := packetcodec.Encode(wireWorldDeltaPacket(WorldDeltaPacket{
		Type: PacketTypeWorldDelta,
		Pickups: FieldRecordDelta[WorldPickupRecord]{
			Updates: []map[string]any{
				{
					"id":          "pickup-1",
					"x":           6,
					"y":           7,
					"age_seconds": 4.5,
				},
			},
		},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	wire := mustDecodeWirePacket(t, encoded)
	updates := mustSliceValue(t, wire, "pickup_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one pickup update, got %#v", updates)
	}

	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "id", "pickup-1")
	assertFloatValue(t, update, "x", 6)
	assertFloatValue(t, update, "y", 7)
	assertFloatValue(t, update, "age_seconds", 4.5)
	assertNotContainsKey(t, update, "type")
	assertNotContainsKey(t, update, "pickup_class")
	assertNotContainsKey(t, update, "health")
	assertNotContainsKey(t, update, "lifespan_seconds")
}

func TestWireSessionDeltaPacketUsesSparseOmission(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneSession,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: SessionWireLaneDelta{
			Metadata: Metadata{Lane: LaneSession, Sequence: 14, SnapshotKind: SnapshotKind("delta")},
		},
	})

	assertStringValue(t, wire, "type", PacketTypeSessionDelta)
	assertStringValue(t, wire, "lane", string(LaneSession))
	assertIntValue(t, wire, "sequence", 14)
	assertStringValue(t, wire, "snapshot_kind", "delta")
	for _, key := range []string{"players", "player_session_updates", "player_session_deletes", "player_lifecycle", "player_lifecycle_updates", "player_lifecycle_deletes", "total_asteroids"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireOverlayWireDeltaPacketOmitsEmptySections(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneOverlay,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: OverlayWireLaneDelta{
			Metadata: Metadata{Lane: LaneOverlay, Sequence: 12, SnapshotKind: SnapshotKind("delta")},
		},
	})

	assertStringValue(t, wire, "type", PacketTypeOverlayDelta)
	assertStringValue(t, wire, "lane", string(LaneOverlay))
	assertIntValue(t, wire, "sequence", 12)
	assertStringValue(t, wire, "snapshot_kind", "delta")
	for _, key := range []string{"receiver_creates", "receiver_updates", "receiver_deletes"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireOverlayWireDeltaPacketKeepsReceiverUpdatesAndOmitsEmptySections(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneOverlay,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: OverlayWireLaneDelta{
			Metadata: Metadata{Lane: LaneOverlay, Sequence: 13, SnapshotKind: SnapshotKind("delta")},
			Receiver: FieldRecordDelta[OverlayReceiverWireRecord]{
				Updates: []map[string]any{{"self_id": "player-1", "score": int64(0), "lives": int64(0), "primary_cooldown_remaining": int64(0)}},
			},
		},
	})

	updates := mustSliceValue(t, wire, "receiver_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one receiver update, got %#v", updates)
	}
	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "self_id", "player-1")
	assertInt64Value(t, update, "score", 0)
	assertInt64Value(t, update, "lives", 0)
	assertInt64Value(t, update, "primary_cooldown_remaining", 0)
	for _, key := range []string{"receiver_creates", "receiver_deletes"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireSessionWireDeltaPacketOmitsEmptySections(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneSession,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: SessionWireLaneDelta{
			Metadata: Metadata{Lane: LaneSession, Sequence: 14, SnapshotKind: SnapshotKind("delta")},
		},
	})

	assertStringValue(t, wire, "type", PacketTypeSessionDelta)
	assertStringValue(t, wire, "lane", string(LaneSession))
	assertIntValue(t, wire, "sequence", 14)
	assertStringValue(t, wire, "snapshot_kind", "delta")
	for _, key := range []string{"players", "player_session_updates", "player_session_deletes", "player_lifecycle", "player_lifecycle_updates", "player_lifecycle_deletes", "total_asteroids"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireSessionWireDeltaPacketKeepsZeroTotalAsteroids(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneSession,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: SessionWireLaneDelta{
			Metadata: Metadata{Lane: LaneSession, Sequence: 15, SnapshotKind: SnapshotKind("delta")},
			TotalAsteroids: RecordDelta[SessionTotalAsteroidsRecord]{Updates: []SessionTotalAsteroidsRecord{{ID: "session-1", Count: 0}}},
		},
	})

	assertIntValue(t, wire, "total_asteroids", 0)
}

func TestWireSessionDeltaPacketEncodesPlayerSessionUpdates(t *testing.T) {
	wire := wireSessionDeltaPacket(SessionLaneDelta{
		Metadata: Metadata{Lane: LaneSession},
		Players: FieldRecordDelta[SessionPlayerRecord]{Updates: []map[string]any{{"id": "player-1", "score": 10}}},
	})

	updates := mustSliceValue(t, wire, "player_session_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one player session update, got %#v", updates)
	}
	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "id", "player-1")
	assertIntValue(t, update, "score", 10)
}

func TestWireSessionDeltaPacketEncodesPlayerLifecycleUpdates(t *testing.T) {
	wire := wireSessionDeltaPacket(SessionLaneDelta{
		Metadata: Metadata{Lane: LaneSession},
		PlayerLifecycle: FieldRecordDelta[SessionLifecycleRecord]{Updates: []map[string]any{{"player_id": "player-1", "status": "respawning"}}},
	})

	updates := mustSliceValue(t, wire, "player_lifecycle_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one player lifecycle update, got %#v", updates)
	}
	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "player_id", "player-1")
	assertStringValue(t, update, "status", "respawning")
}

func TestWireSessionDeltaPacketEncodesPlayerLifecycleDeletes(t *testing.T) {
	wire := wireSessionDeltaPacket(SessionLaneDelta{
		Metadata: Metadata{Lane: LaneSession},
		PlayerLifecycle: FieldRecordDelta[SessionLifecycleRecord]{Creates: []SessionLifecycleRecord{{PlayerID: "player-1", Status: "active"}}, Updates: []map[string]any{{"player_id": "player-1", "status": "respawning"}}, Deletes: []string{"player-1"}},
	})

	deletes := wire["player_lifecycle_deletes"]
	items, ok := deletes.([]string)
	if !ok {
		t.Fatalf("expected player_lifecycle_deletes to be a string array, got %#v", deletes)
	}
	if len(items) != 1 || items[0] != "player-1" {
		t.Fatalf("expected one player lifecycle delete, got %#v", items)
	}
}

func TestActiveWirePacketEncodingUsesWorldDeltaEnvelope(t *testing.T) {
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
			Ships: FieldRecordDelta[WorldShipRecord]{Creates: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing"}}, Updates: []map[string]any{{"id": "ship-a", "x": 2}}, Deletes: []string{"ship-b"}},
			Bullets: FieldRecordDelta[WorldBulletRecord]{Creates: []WorldBulletRecord{{ID: "bullet-a", X: 1, Y: 2, Rotation: 3, OwnerID: "ship-a", WeaponID: "pulse", ProjectileType: "laser"}}, Updates: []map[string]any{{"id": "bullet-a", "x": 4, "y": 5}}, Deletes: []string{"bullet-z"}},
			Asteroids: FieldRecordDelta[WorldAsteroidRecord]{Creates: []WorldAsteroidRecord{{ID: "asteroid-a", X: 1, Y: 2, Size: 3, Health: 4, Scale: 5, Variant: 1}}, Updates: []map[string]any{{"id": "asteroid-a", "x": 6}}, Deletes: []string{"asteroid-a"}},
			Pickups: FieldRecordDelta[WorldPickupRecord]{Creates: []WorldPickupRecord{{ID: "pickup-a", Type: "shield", PickupClass: "powerup", X: 1, Y: 2, Health: 3, AgeSeconds: 4, LifespanSeconds: 5}}, Updates: []map[string]any{{"id": "pickup-a", "x": 7}}, Deletes: []string{"pickup-a"}},
		},
	}

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketTypeWorldDelta)
	assertStringValue(t, wire, "lane", string(LaneWorld))
	assertIntValue(t, wire, "sequence", 9)
	assertStringValue(t, wire, "baseline_id", "baseline-9")
	assertStringValue(t, wire, "snapshot_id", "snapshot-9")
	assertStringValue(t, wire, "snapshot_kind", "delta")
	assertContainsKey(t, wire, "ship_creates")
	assertContainsKey(t, wire, "bullet_updates")
	assertContainsKey(t, wire, "asteroid_deletes")
	assertContainsKey(t, wire, "pickup_creates")
	assertNotNakedDeltaPayload(t, wire)
}

func TestActiveWirePacketEncodingUsesOverlayDeltaEnvelope(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneOverlay,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: OverlayLaneDelta{
			Metadata: Metadata{
				Lane:         LaneOverlay,
				Sequence:     12,
				BaselineID:   "overlay-baseline-12",
				SnapshotID:   "overlay-snapshot-12",
				SnapshotKind: SnapshotKind("delta"),
			},
			Receiver: FieldRecordDelta[OverlayReceiverRecord]{Updates: []map[string]any{{"self_id": "player-1", "score": 10, "primary_cooldown_remaining": 1.25}}},
		},
	}

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketTypeOverlayDelta)
	assertStringValue(t, wire, "lane", string(LaneOverlay))
	assertIntValue(t, wire, "sequence", 12)
	assertStringValue(t, wire, "baseline_id", "overlay-baseline-12")
	assertStringValue(t, wire, "snapshot_id", "overlay-snapshot-12")
	assertStringValue(t, wire, "snapshot_kind", "delta")
	assertContainsKey(t, wire, "receiver_updates")
	assertNotNakedOverlayDeltaPayload(t, wire)
}

func TestWireOverlayDeltaPacketEncodesReceiverUpdatesAsPartialFieldPatch(t *testing.T) {
	encoded, err := packetcodec.Encode(wireOverlayDeltaPacket(OverlayLaneDelta{
		Metadata: Metadata{Lane: LaneOverlay, Sequence: 12, BaselineID: "overlay-baseline-12", SnapshotID: "overlay-snapshot-12", SnapshotKind: SnapshotKind("delta")},
		Receiver: FieldRecordDelta[OverlayReceiverRecord]{Updates: []map[string]any{{"self_id": "player-1", "score": 10, "primary_cooldown_remaining": 1.25}}},
	}))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	wire := mustDecodeWirePacket(t, encoded)
	assertStringValue(t, wire, "type", PacketTypeOverlayDelta)
	assertStringValue(t, wire, "lane", string(LaneOverlay))
	assertIntValue(t, wire, "sequence", 12)
	assertStringValue(t, wire, "baseline_id", "overlay-baseline-12")
	assertStringValue(t, wire, "snapshot_id", "overlay-snapshot-12")
	assertStringValue(t, wire, "snapshot_kind", "delta")

	updates := mustSliceValue(t, wire, "receiver_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one receiver update, got %#v", updates)
	}
	update := mustMapValue(t, updates[0])
	assertStringValue(t, update, "self_id", "player-1")
	assertIntValue(t, update, "score", 10)
	assertFloatValue(t, update, "primary_cooldown_remaining", 1.25)
	assertNotContainsKey(t, update, "lives")
	assertNotContainsKey(t, update, "primary_weapon_id")
	assertNotContainsKey(t, update, "secondary_weapon_id")
	assertNotContainsKey(t, update, "primary_ammo_policy")
	assertNotContainsKey(t, update, "secondary_ammo_policy")
	assertNotContainsKey(t, update, "primary_ammo_remaining")
	assertNotContainsKey(t, update, "secondary_ammo_remaining")
}

func TestWireOverlayWireFullPacketEncodesIntegerCooldownFields(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneOverlay,
		Kind: RealtimeLaneCandidateKindFull,
		Full: OverlayWireFullPacket{
			Type: PacketFamilyOverlayFull,
			Metadata: Metadata{Lane: LaneOverlay, Sequence: 3},
			Receiver: OverlayReceiverWireRecord{
			SelfID:                    "player-1",
			Lives:                     2,
			Score:                     9,
			RespawnCooldown:           1250,
			PrimaryWeaponID:           "pulse",
			PrimaryAmmoPolicy:         "limited",
			PrimaryCooldownRemaining:   500,
			PrimaryAmmoRemaining:      12,
			SecondaryWeaponID:         "mine",
			SecondaryAmmoPolicy:       "infinite",
			SecondaryCooldownRemaining: 750,
			SecondaryAmmoRemaining:    3,
			},
		},
	})

	assertInt64Value(t, wire, "respawn_cooldown", 1250)
	assertInt64Value(t, wire, "primary_cooldown_remaining", 500)
	assertInt64Value(t, wire, "secondary_cooldown_remaining", 750)
}

func TestWireOverlayWireDeltaPacketEncodesIntegerCooldownUpdates(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneOverlay,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: OverlayWireLaneDelta{
			Metadata: Metadata{Lane: LaneOverlay, Sequence: 12, BaselineID: "overlay-baseline-12", SnapshotID: "overlay-snapshot-12", SnapshotKind: SnapshotKind("delta")},
			Receiver: FieldRecordDelta[OverlayReceiverWireRecord]{Updates: []map[string]any{{"self_id": "player-1", "primary_cooldown_remaining": int64(500)}}},
		},
	})

	updates := mustSliceValue(t, wire, "receiver_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one receiver update, got %#v", updates)
	}
	update := mustMapValue(t, updates[0])
	assertInt64Value(t, update, "primary_cooldown_remaining", 500)
}

func TestActiveWirePacketEncodingUsesSessionDeltaEnvelope(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneSession,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: SessionLaneDelta{
			Metadata: Metadata{
				Lane:         LaneSession,
				Sequence:     14,
				BaselineID:   "session-baseline-14",
				SnapshotID:   "session-snapshot-14",
				SnapshotKind: SnapshotKind("delta"),
			},
			Players: FieldRecordDelta[SessionPlayerRecord]{Updates: []map[string]any{{"id": "player-1", "score": 10, "lives": 2}}},
			PlayerLifecycle: FieldRecordDelta[SessionLifecycleRecord]{Updates: []map[string]any{{"player_id": "player-1", "status": "respawning"}}},
			TotalAsteroids: RecordDelta[SessionTotalAsteroidsRecord]{Updates: []SessionTotalAsteroidsRecord{{ID: "session-14", Count: 8}}},
		},
	}

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketTypeSessionDelta)
	assertStringValue(t, wire, "lane", string(LaneSession))
	assertIntValue(t, wire, "sequence", 14)
	assertStringValue(t, wire, "baseline_id", "session-baseline-14")
	assertStringValue(t, wire, "snapshot_id", "session-snapshot-14")
	assertStringValue(t, wire, "snapshot_kind", "delta")
	assertNotContainsKey(t, wire, "players")
	assertContainsKey(t, wire, "player_session_updates")
	assertNotContainsKey(t, wire, "player_lifecycle")
	assertContainsKey(t, wire, "player_lifecycle_updates")
	assertContainsKey(t, wire, "total_asteroids")
	assertNotNakedSessionDeltaPayload(t, wire)
}

func TestActiveWirePacketEncodingUsesLowercaseOverlayShape(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneOverlay,
		Kind: RealtimeLaneCandidateKindFull,
		Full: OverlayFullPacket{
			Type: PacketFamilyOverlayFull,
			Metadata: Metadata{
				Lane:     LaneOverlay,
				Sequence: 3,
			},
			Receiver: OverlayReceiverRecord{
				SelfID:                   "player-1",
				Lives:                    2,
				Score:                    9,
				RespawnCooldown:          1.25,
				PrimaryWeaponID:          "pulse",
				PrimaryAmmoPolicy:        "limited",
				PrimaryCooldownRemaining: 0.5,
				PrimaryAmmoRemaining:     12,
				SecondaryWeaponID:        "mine",
				SecondaryAmmoPolicy:      "infinite",
				SecondaryCooldownRemaining: 0.75,
				SecondaryAmmoRemaining:   3,
			},
		},
	}

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilyOverlayFull)
	assertStringValue(t, wire, "self_id", "player-1")
	assertContainsKey(t, wire, "respawn_cooldown")
	assertNotContainsKey(t, wire, "respawn")
	assertNotContainsKey(t, wire, "Receiver")
}

func TestActiveWirePacketEncodingUsesLowercaseSessionShape(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneSession,
		Kind: RealtimeLaneCandidateKindFull,
		Full: SessionFullPacket{
			Type: PacketFamilySessionFull,
			Metadata: Metadata{
				Lane:     LaneSession,
				Sequence: 5,
			},
			Players: []SessionPlayerRecord{
				{
					ID:                  "player-1",
					ShipType:            "v_wing",
					Score:               8,
					Lives:               3,
					RespawnCooldown:     0.25,
					PrimaryWeaponID:     "pulse",
					PrimaryAmmoPolicy:   "limited",
					SecondaryWeaponID:   "mine",
					SecondaryAmmoPolicy: "infinite",
					SpawnX:              10,
					SpawnY:              20,
				},
			},
			PlayerLifecycle: []SessionLifecycleRecord{
				{
					PlayerID: "player-1",
					Status:   "active",
				},
			},
			TotalAsteroids: 42,
		},
	}

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilySessionFull)
	assertContainsKey(t, wire, "players")
	assertContainsKey(t, wire, "player_lifecycle")
	assertIntValue(t, wire, "total_asteroids", 42)
}

func TestActiveWirePacketEncodingUsesLowercaseEventShape(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneEvent,
		Kind: RealtimeLaneCandidateKindEventBatch,
		Full: EventBatchPacket{
			Type: PacketFamilyEventBatch,
			Metadata: Metadata{
				Lane:     LaneEvent,
				Sequence: 11,
			},
			Batch: EventBatchRecord{
				BatchID:  "event-batch-11",
				Sequence: 11,
				Events: []EventRecord{
					{
						EventID: "event-1",
						Event: game.EventState{
							Type:       "bullet_blast",
							X:          1,
							Y:          2,
							SourceID:   "ship-1",
							EffectType: "blast",
						},
					},
					{
						EventID: "event-2",
						Event: game.EventState{
							Type:         "ship_death",
							PlayerID:     "player-1",
							Lives:        2,
							RespawnDelay: 3.5,
							X:            4,
							Y:            5,
							SourceID:     "ship-2",
							EffectType:   "death",
						},
					},
				},
			},
		},
	}

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilyEventBatch)
	assertContainsKey(t, wire, "events")
	events := mustSliceValue(t, wire, "events")
	bulletBlast := mustMapValue(t, events[0])
	assertStringValue(t, bulletBlast, "event_id", "event-1")
	assertStringValue(t, bulletBlast, "type", "bullet_blast")
	assertFloatValue(t, bulletBlast, "x", 1)
	assertFloatValue(t, bulletBlast, "y", 2)
	assertStringValue(t, bulletBlast, "source_id", "ship-1")
	assertStringValue(t, bulletBlast, "effect_type", "blast")

	shipDeath := mustMapValue(t, events[1])
	assertStringValue(t, shipDeath, "event_id", "event-2")
	assertStringValue(t, shipDeath, "type", "ship_death")
	assertStringValue(t, shipDeath, "player_id", "player-1")
	assertIntValue(t, shipDeath, "lives", 2)
	assertFloatValue(t, shipDeath, "respawn_delay", 3.5)
	assertFloatValue(t, shipDeath, "x", 4)
	assertFloatValue(t, shipDeath, "y", 5)
	assertStringValue(t, shipDeath, "source_id", "ship-2")
	assertStringValue(t, shipDeath, "effect_type", "death")
}

func TestWireLanePacketDropsUnsupportedFullPayloads(t *testing.T) {
	if wire := WireLanePacket(RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull, Full: map[string]any{"type": "world_full"}}); len(wire) != 0 {
		t.Fatalf("expected unsupported full map payload to be dropped, got %#v", wire)
	}
	if wire := WireLanePacket(RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindFull, Full: struct{ Type string }{Type: "world_full"}}); len(wire) != 0 {
		t.Fatalf("expected unsupported full struct payload to be dropped, got %#v", wire)
	}
}

func TestWireLanePacketDropsUnsupportedDeltaPayloads(t *testing.T) {
	if wire := WireLanePacket(RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindDelta, Delta: map[string]any{"ship_creates": []any{}}}); len(wire) != 0 {
		t.Fatalf("expected unsupported delta map payload to be dropped, got %#v", wire)
	}
	if wire := WireLanePacket(RealtimeLaneCandidate{Lane: LaneWorld, Kind: RealtimeLaneCandidateKindDelta, Delta: struct{ ShipCreates []any }{ShipCreates: []any{}}}); len(wire) != 0 {
		t.Fatalf("expected unsupported delta struct payload to be dropped, got %#v", wire)
	}
}

func mustEncodeWirePacket(t *testing.T, candidate RealtimeLaneCandidate) []byte {
	t.Helper()

	encoded, err := packetcodec.Encode(WireLanePacket(candidate))
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	return encoded
}

func mustDecodeWirePacket(t *testing.T, encoded []byte) map[string]any {
	t.Helper()

	var wire map[string]any
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	return wire
}

func mustMapValue(t *testing.T, value any) map[string]any {
	t.Helper()

	wire, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected map value, got %#v", value)
	}
	return wire
}

func mustSliceValue(t *testing.T, wire map[string]any, key string) []any {
	t.Helper()

	value, ok := wire[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("expected key %q to be an array, got %#v", key, value)
	}
	return items
}

func assertStringValue(t *testing.T, wire map[string]any, key string, want string) {
	t.Helper()

	value, ok := wire[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}
	got, ok := value.(string)
	if !ok {
		t.Fatalf("expected key %q to be a string, got %#v", key, value)
	}
	if got != want {
		t.Fatalf("key %q = %q, want %q", key, got, want)
	}
}

func assertFloatValue(t *testing.T, wire map[string]any, key string, want float64) {
	t.Helper()

	value, ok := wire[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}
	got, ok := value.(float64)
	if !ok {
		t.Fatalf("expected key %q to be numeric, got %#v", key, value)
	}
	if got != want {
		t.Fatalf("key %q = %v, want %v", key, got, want)
	}
}

func assertInt64Value(t *testing.T, wire map[string]any, key string, want int64) {
	t.Helper()

	value, ok := wire[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}
	got, ok := value.(int64)
	if !ok {
		t.Fatalf("expected key %q to be int64, got %#v", key, value)
	}
	if got != want {
		t.Fatalf("key %q = %v, want %d", key, got, want)
	}
}

func assertJSONIntValue(t *testing.T, wire map[string]any, key string, want int64) {
	t.Helper()

	value, ok := wire[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}
	got, ok := value.(float64)
	if !ok {
		t.Fatalf("expected key %q to be numeric, got %#v", key, value)
	}
	if got != float64(want) {
		t.Fatalf("key %q = %v, want %d", key, got, want)
	}
	if got != float64(int64(got)) {
		t.Fatalf("expected key %q to be an integer value, got %v", key, got)
	}
}

func assertIntValue(t *testing.T, wire map[string]any, key string, want int) {
	t.Helper()

	value, ok := wire[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}

	var got int
	switch typed := value.(type) {
	case int:
		got = typed
	case int64:
		got = int(typed)
	case float64:
		got = int(typed)
	default:
		t.Fatalf("expected key %q to be numeric, got %#v", key, value)
	}

	if got != want {
		t.Fatalf("key %q = %v, want %d", key, value, want)
	}
}

func assertContainsKey(t *testing.T, wire map[string]any, key string) {
	t.Helper()
	if _, ok := wire[key]; !ok {
		t.Fatalf("expected key %q to exist", key)
	}
}

func assertNotContainsKey(t *testing.T, wire map[string]any, key string) {
	t.Helper()
	for existingKey := range wire {
		if existingKey == key {
			t.Fatalf("did not expect key %q", key)
		}
	}
}
func TestWireLanePacketRoundTripsWorldFullFamily(t *testing.T) {
	candidate := RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindFull,
		Full: WorldFullPacket{
			Type: PacketFamilyWorldFull,
			Metadata: Metadata{Lane: LaneWorld, Sequence: 21},
			Ships: []WorldShipRecord{{ID: "ship-1", ShipType: "v_wing", X: 1, Y: 2, Rotation: 3, Health: 4, Shields: 5, Thrusting: true, TargetKind: "player", TargetID: "player-1"}},
			Bullets: []WorldBulletRecord{{ID: "bullet-1", OwnerID: "ship-1", X: 6, Y: 7, Rotation: 8, WeaponID: "pulse", ProjectileType: "laser"}},
			Asteroids: []WorldAsteroidRecord{{ID: "asteroid-1", X: 9, Y: 10, Size: 2, Health: 11, Scale: 1.5, Variant: 3}},
			Pickups: []WorldPickupRecord{{ID: "pickup-1", Type: "shield", PickupClass: "armor", X: 12, Y: 13, Health: 1, AgeSeconds: 4.5, LifespanSeconds: 9.5}},
		},
	}

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilyWorldFull)
	assertContainsKey(t, wire, "ships")
	assertContainsKey(t, wire, "bullets")
	assertContainsKey(t, wire, "asteroids")
	assertContainsKey(t, wire, "pickups")
}

func TestWireSessionWireFullPacketEncodesIntegerCooldownFields(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneSession,
		Kind: RealtimeLaneCandidateKindFull,
		Full: SessionWireFullPacket{
			Type: PacketFamilySessionFull,
			Metadata: Metadata{Lane: LaneSession, Sequence: 5},
			Players: []SessionPlayerWireRecord{{
				ID:                  "player-1",
				ShipType:            "v_wing",
				Score:               8,
				Lives:               3,
				RespawnCooldown:     250,
				PrimaryWeaponID:     "pulse",
				PrimaryAmmoPolicy:   "limited",
				SecondaryWeaponID:   "mine",
				SecondaryAmmoPolicy: "infinite",
				SpawnX:              10,
				SpawnY:              20,
			}},
			PlayerLifecycle: []SessionLifecycleRecord{{PlayerID: "player-1", Status: "active"}},
			TotalAsteroids: 42,
		},
	})


	players := mustSliceValue(t, wire, "players")
	if len(players) != 1 {
		t.Fatalf("expected one player, got %#v", players)
	}
	player := mustMapValue(t, players[0])
	assertInt64Value(t, player, "respawn_cooldown", 250)
	assertInt64Value(t, player, "spawn_x", 10)
	assertInt64Value(t, player, "spawn_y", 20)
}
func TestWireLanePacketContainsLowercaseKeysOnly(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindFull,
		Full: WorldFullPacket{Type: PacketFamilyWorldFull},
	})

	for key := range wire {
		if strings.ToLower(key) != key {
			t.Fatalf("expected lowercase key, got %q", key)
		}
	}
}


func assertNotNakedDeltaPayload(t *testing.T, wire map[string]any) {
	t.Helper()
	if hasOnlyKeys(wire, []string{"ship_creates", "ship_updates", "ship_deletes", "bullet_creates", "bullet_updates", "bullet_deletes", "asteroid_creates", "asteroid_updates", "asteroid_deletes", "pickup_creates", "pickup_updates", "pickup_deletes"}) {
		t.Fatalf("world delta payload encoded without envelope: %#v", wire)
	}
}

func assertNotNakedOverlayDeltaPayload(t *testing.T, wire map[string]any) {
	t.Helper()
	if hasOnlyKeys(wire, []string{"receiver_creates", "receiver_updates", "receiver_deletes"}) {
		t.Fatalf("overlay delta payload encoded without envelope: %#v", wire)
	}
}

func assertNotNakedSessionDeltaPayload(t *testing.T, wire map[string]any) {
	t.Helper()
	if hasOnlyKeys(wire, []string{"players", "player_session_updates", "player_session_deletes", "player_lifecycle", "player_lifecycle_updates", "player_lifecycle_deletes", "total_asteroids"}) {
		t.Fatalf("session delta payload encoded without envelope: %#v", wire)
	}
}


func TestCandidateMetadataReturnsWorldWirePacketMetadata(t *testing.T) {
	full := WorldWireFullPacket{Type: PacketFamilyWorldFull, Metadata: Metadata{Lane: LaneWorld, Sequence: 21, SnapshotKind: SnapshotKind("full")}}
	fullMetadata, ok := CandidateMetadata(RealtimeLaneCandidate{Lane: LaneWorld, Full: full}, NewRealtimeSessionState("player-1"))
	if !ok {
		t.Fatal("expected world wire full metadata to be found")
	}
	if fullMetadata != full.Metadata {
		t.Fatalf("full metadata = %#v, want %#v", fullMetadata, full.Metadata)
	}

	delta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 22, SnapshotKind: SnapshotKind("delta")}}
	deltaMetadata, ok := CandidateMetadata(RealtimeLaneCandidate{Lane: LaneWorld, Delta: delta}, NewRealtimeSessionState("player-1"))
	if !ok {
		t.Fatal("expected world wire delta metadata to be found")
	}
	if deltaMetadata != delta.Metadata {
		t.Fatalf("delta metadata = %#v, want %#v", deltaMetadata, delta.Metadata)
	}
}

func TestWireWorldWireFullPacketEncodesIntegerWorldFields(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindFull,
		Full: WorldWireFullPacket{
			Type: PacketFamilyWorldFull,
			Metadata: Metadata{Lane: LaneWorld, Sequence: 7},
			Ships: []WorldShipWireRecord{{ID: "ship-1", ShipType: "v_wing", X: 10, Y: 20, Rotation: 30, Health: 4, Shields: 5, Thrusting: true, TargetKind: "player", TargetID: "player-1"}},
			Bullets: []WorldBulletWireRecord{{ID: "bullet-1", OwnerID: "ship-1", X: 6, Y: 7, Rotation: 8, WeaponID: "pulse", ProjectileType: "laser"}},
			Asteroids: []WorldAsteroidWireRecord{{ID: "asteroid-1", X: 9, Y: 10, Size: 2, Health: 11, Scale: 15, Variant: 3}},
			Pickups: []WorldPickupWireRecord{{ID: "pickup-1", Type: "shield", PickupClass: "armor", X: 12, Y: 13, Health: 1, AgeSeconds: 4, LifespanSeconds: 9}},
		},
	})

	ships := mustSliceValue(t, wire, "ships")
	ship := mustMapValue(t, ships[0])
	assertInt64Value(t, ship, "x", 10)
	assertInt64Value(t, ship, "y", 20)
	assertInt64Value(t, ship, "rotation", 30)

	asteroids := mustSliceValue(t, wire, "asteroids")
	asteroid := mustMapValue(t, asteroids[0])
	assertInt64Value(t, asteroid, "scale", 15)
}

func TestWireWorldWireDeltaPacketEncodesIntegerWorldFieldUpdates(t *testing.T) {
	wire := WireLanePacket(RealtimeLaneCandidate{
		Lane: LaneWorld,
		Kind: RealtimeLaneCandidateKindDelta,
		Delta: WorldWireDeltaPacket{
			Type: PacketTypeWorldDelta,
			Metadata: Metadata{Lane: LaneWorld, Sequence: 9, SnapshotKind: SnapshotKind("delta")},
			Ships: FieldRecordDelta[WorldShipWireRecord]{Updates: []map[string]any{{"id": "ship-1", "x": int64(10), "y": int64(20), "rotation": int64(30), "thrusting": true}}},
		},
	})

	updates := mustSliceValue(t, wire, "ship_updates")
	update := mustMapValue(t, updates[0])
	assertInt64Value(t, update, "x", 10)
	assertInt64Value(t, update, "y", 20)
	assertInt64Value(t, update, "rotation", 30)
	assertNotContainsKey(t, update, "ship_type")
}

