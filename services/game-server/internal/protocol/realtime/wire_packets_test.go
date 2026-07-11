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
	candidate := mustRealtimeLaneCandidate(WorldFullPacket{
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
	}, nil)

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

	encoded, err := packetcodec.Encode(mustWireLanePacket(t, mustRealtimeLaneCandidate(wire, nil)))
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
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, mustRealtimeLaneCandidate(WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
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
	}, nil)))

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
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, mustRealtimeLaneCandidate(WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 10, SnapshotKind: SnapshotKind("delta")},
		Bullets: FieldRecordDelta[WorldBulletWireRecord]{
			Deletes: []string{"bullet-1"},
		},
	}, nil)))

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
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, mustRealtimeLaneCandidate(WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
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
	}, nil)))

	update := mustMapValue(t, mustSliceValue(t, wire, "ship_updates")[0])
	assertJSONIntValue(t, update, "x", 0)
	assertJSONIntValue(t, update, "y", 0)
	assertJSONIntValue(t, update, "rotation", 0)
	if thrusting, ok := update["thrusting"].(bool); !ok || thrusting {
		t.Fatalf("expected thrusting to be false, got %#v", update["thrusting"])
	}
}

func TestWireWorldWireDeltaPacketSparseJsonOmitsEmptySections(t *testing.T) {
	encoded, err := packetcodec.Encode(mustWireLanePacket(t, mustRealtimeLaneCandidate(WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 20, SnapshotKind: SnapshotKind("delta")},
		Ships: FieldRecordDelta[WorldShipWireRecord]{
			Updates: []map[string]any{{"id": "ship-1", "x": int64(1), "y": int64(2), "rotation": int64(3), "thrusting": true}},
		},
	}, nil)))
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
	encoded, err := packetcodec.Encode(mustWireLanePacket(t, mustRealtimeLaneCandidate(OverlayWireLaneDelta{
		Metadata: Metadata{Lane: LaneOverlay, Sequence: 21, SnapshotKind: SnapshotKind("delta")},
	}, nil)))
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
	encoded, err := packetcodec.Encode(mustWireLanePacket(t, mustRealtimeLaneCandidate(SessionWireLaneDelta{
		Metadata: Metadata{Lane: LaneSession, Sequence: 22, SnapshotKind: SnapshotKind("delta")},
	}, nil)))
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
					"id":        "ship-1",
					"x":         6,
					"y":         7,
					"rotation":  8,
					"thrusting": true,
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

func TestWireAsteroidDeltaPacketIsUpdateOnly(t *testing.T) {
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{
		Type:     PacketFamilyAsteroidDelta,
		Metadata: Metadata{Lane: LaneAsteroids, Sequence: 42, ServerSentMsec: 123456, SnapshotKind: SnapshotKind("delta"), ChunkIndex: 1, ChunkCount: 3, IsFinalChunk: false},
		AsteroidUpdates: []map[string]any{
			{
				"id":     "asteroid-1",
				"x":      int64(10),
				"y":      int64(20),
				"size":   int64(3),
				"health": int64(4),
			},
		},
	}, nil)))

	assertStringValue(t, wire, "type", PacketFamilyAsteroidDelta)
	assertIntValue(t, wire, "sequence", 42)
	assertIntValue(t, wire, "server_sent_msec", 123456)
	assertIntValue(t, wire, "chunk_index", 1)
	assertIntValue(t, wire, "chunk_count", 3)
	assertNotContainsKey(t, wire, "is_final_chunk")
	assertContainsKey(t, wire, "asteroid_updates")
	for _, key := range []string{"asteroid_creates", "asteroid_deletes", "bullet_creates", "bullet_deletes"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireAsteroidLifecyclePacketEncodesCreatesAndDeletes(t *testing.T) {
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, mustRealtimeLaneCandidate(AsteroidWireDeltaPacket{
		Type: PacketFamilyAsteroidsLifecycle, Metadata: Metadata{Lane: LaneAsteroidsLifecycle, Sequence: 44, BaselineID: "world-baseline-44", SnapshotID: "world-snapshot-44", ServerSentMsec: 987654, SnapshotKind: SnapshotKind("delta"), ChunkIndex: 0, ChunkCount: 1, IsFinalChunk: true},
		AsteroidCreates: []WorldAsteroidWireRecord{{ID: "asteroid-1", X: 10, Y: 20, Size: 3, Health: 4, Scale: 5, Variant: 6}},
		AsteroidDeletes: []string{"asteroid-9"},
	}, nil)))

	assertStringValue(t, wire, "type", PacketFamilyAsteroidsLifecycle)
	assertStringValue(t, wire, "lane", string(LaneAsteroidsLifecycle))
	assertIntValue(t, wire, "sequence", 44)
	assertStringValue(t, wire, "baseline_id", "world-baseline-44")
	assertStringValue(t, wire, "snapshot_id", "world-snapshot-44")
	assertStringValue(t, wire, "snapshot_kind", "delta")
	assertIntValue(t, wire, "server_sent_msec", 987654)
	assertContainsKey(t, wire, "asteroid_creates")
	assertContainsKey(t, wire, "asteroid_deletes")
	assertNotContainsKey(t, wire, "asteroid_updates")
	creates := mustSliceValue(t, wire, "asteroid_creates")
	if len(creates) != 1 {
		t.Fatalf("expected one asteroid create, got %#v", creates)
	}
	create := mustMapValue(t, creates[0])
	assertStringValue(t, create, "id", "asteroid-1")
	assertFloatValue(t, create, "x", 10)
	assertFloatValue(t, create, "y", 20)
	assertFloatValue(t, create, "size", 3)
	assertFloatValue(t, create, "health", 4)
	assertFloatValue(t, create, "scale", 5)
	assertFloatValue(t, create, "variant", 6)
	deletes := mustSliceValue(t, wire, "asteroid_deletes")
	if len(deletes) != 1 || deletes[0] != "asteroid-9" {
		t.Fatalf("expected one asteroid delete, got %#v", deletes)
	}
}

func TestWireBulletDeltaPacketIsUpdateOnly(t *testing.T) {
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, mustRealtimeLaneCandidate(BulletWireDeltaPacket{
		Type:     PacketFamilyBulletDelta,
		Metadata: Metadata{Lane: LaneBullets, Sequence: 43, ServerSentMsec: 654321, SnapshotKind: SnapshotKind("delta"), ChunkIndex: 2, ChunkCount: 4, IsFinalChunk: true},
		BulletUpdates: []map[string]any{
			{
				"id":       "bullet-1",
				"x":        int64(30),
				"y":        int64(40),
				"rotation": int64(50),
			},
		},
	}, nil)))

	assertStringValue(t, wire, "type", PacketFamilyBulletDelta)
	assertIntValue(t, wire, "sequence", 43)
	assertIntValue(t, wire, "server_sent_msec", 654321)
	assertIntValue(t, wire, "chunk_index", 2)
	assertIntValue(t, wire, "chunk_count", 4)
	assertContainsKey(t, wire, "is_final_chunk")
	assertContainsKey(t, wire, "bullet_updates")
	for _, key := range []string{"bullet_creates", "bullet_deletes", "asteroid_creates", "asteroid_deletes"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireBulletLifecyclePacketEncodesCreatesAndDeletes(t *testing.T) {
	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, mustRealtimeLaneCandidate(BulletWireDeltaPacket{
		Type:          PacketFamilyBulletsLifecycle,
		Metadata:      Metadata{Lane: LaneBulletsLifecycle, Sequence: 45, BaselineID: "world-baseline-45", SnapshotID: "world-snapshot-45", ServerSentMsec: 123, SnapshotKind: SnapshotKind("delta"), ChunkIndex: 0, ChunkCount: 1, IsFinalChunk: true},
		BulletCreates: []WorldBulletWireRecord{{ID: "bullet-1", OwnerID: "ship-1", X: 1, Y: 2, Rotation: 3, WeaponID: "pulse", ProjectileType: "laser"}},
		BulletDeletes: []string{"bullet-9"},
	}, nil)))

	assertStringValue(t, wire, "type", PacketFamilyBulletsLifecycle)
	assertStringValue(t, wire, "lane", string(LaneBulletsLifecycle))
	assertIntValue(t, wire, "sequence", 45)
	assertStringValue(t, wire, "baseline_id", "world-baseline-45")
	assertStringValue(t, wire, "snapshot_id", "world-snapshot-45")
	assertStringValue(t, wire, "snapshot_kind", "delta")
	assertContainsKey(t, wire, "bullet_creates")
	assertContainsKey(t, wire, "bullet_deletes")
	assertNotContainsKey(t, wire, "bullet_updates")
	assertNotContainsKey(t, wire, "asteroid_creates")
	assertNotContainsKey(t, wire, "asteroid_deletes")
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
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(SessionWireLaneDelta{
		Metadata: Metadata{Lane: LaneSession, Sequence: 14, SnapshotKind: SnapshotKind("delta")},
	}, nil))

	assertStringValue(t, wire, "type", PacketTypeSessionDelta)
	assertIntValue(t, wire, "sequence", 14)
	for _, key := range []string{"players", "player_session_updates", "player_session_deletes", "player_lifecycle", "player_lifecycle_updates", "player_lifecycle_deletes", "total_asteroids"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireOverlayWireDeltaPacketOmitsEmptySections(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(OverlayWireLaneDelta{
		Metadata: Metadata{Lane: LaneOverlay, Sequence: 12, SnapshotKind: SnapshotKind("delta")},
	}, nil))

	assertStringValue(t, wire, "type", PacketTypeOverlayDelta)
	assertIntValue(t, wire, "sequence", 12)
	for _, key := range []string{"receiver_creates", "receiver_updates", "receiver_deletes"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireOverlayWireDeltaPacketKeepsReceiverUpdatesAndOmitsEmptySections(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(OverlayWireLaneDelta{
		Metadata: Metadata{Lane: LaneOverlay, Sequence: 13, SnapshotKind: SnapshotKind("delta")},
		Receiver: FieldRecordDelta[OverlayReceiverWireRecord]{
			Updates: []map[string]any{{"self_id": "player-1", "score": int64(0), "lives": int64(0), "primary_cooldown_remaining": int64(0)}},
		},
	}, nil))

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
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(SessionWireLaneDelta{
		Metadata: Metadata{Lane: LaneSession, Sequence: 14, SnapshotKind: SnapshotKind("delta")},
	}, nil))

	assertStringValue(t, wire, "type", PacketTypeSessionDelta)
	assertIntValue(t, wire, "sequence", 14)
	for _, key := range []string{"players", "player_session_updates", "player_session_deletes", "player_lifecycle", "player_lifecycle_updates", "player_lifecycle_deletes", "total_asteroids"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireSessionWireDeltaPacketKeepsZeroTotalAsteroids(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(SessionWireLaneDelta{
		Metadata:       Metadata{Lane: LaneSession, Sequence: 15, SnapshotKind: SnapshotKind("delta")},
		TotalAsteroids: RecordDelta[SessionTotalAsteroidsRecord]{Updates: []SessionTotalAsteroidsRecord{{ID: "session-1", Count: 0}}},
	}, nil))

	assertIntValue(t, wire, "total_asteroids", 0)
}

func TestWireSessionDeltaPacketEncodesPlayerSessionUpdates(t *testing.T) {
	wire := wireSessionDeltaPacket(SessionLaneDelta{
		Metadata: Metadata{Lane: LaneSession},
		Players:  FieldRecordDelta[SessionPlayerRecord]{Updates: []map[string]any{{"id": "player-1", "score": 10}}},
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
		Metadata:        Metadata{Lane: LaneSession},
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
		Metadata:        Metadata{Lane: LaneSession},
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
	candidate := mustRealtimeLaneCandidate(WorldDeltaPacket{
		Type: PacketTypeWorldDelta,
		Metadata: Metadata{
			Lane:         LaneWorld,
			Sequence:     9,
			BaselineID:   "baseline-9",
			SnapshotID:   "snapshot-9",
			SnapshotKind: SnapshotKind("delta"),
		},
		Ships:     FieldRecordDelta[WorldShipRecord]{Creates: []WorldShipRecord{{ID: "ship-a", ShipType: "v_wing"}}, Updates: []map[string]any{{"id": "ship-a", "x": 2}}, Deletes: []string{"ship-b"}},
		Bullets:   FieldRecordDelta[WorldBulletRecord]{Updates: []map[string]any{{"id": "bullet-a", "x": 4, "y": 5}}},
		Asteroids: FieldRecordDelta[WorldAsteroidRecord]{Updates: []map[string]any{{"id": "asteroid-a", "x": 6}}},
		Pickups:   FieldRecordDelta[WorldPickupRecord]{Creates: []WorldPickupRecord{{ID: "pickup-a", Type: "shield", PickupClass: "powerup", X: 1, Y: 2, Health: 3, AgeSeconds: 4, LifespanSeconds: 5}}, Updates: []map[string]any{{"id": "pickup-a", "x": 7}}, Deletes: []string{"pickup-a"}},
	}, nil)

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketTypeWorldDelta)
	assertIntValue(t, wire, "sequence", 9)
	assertStringValue(t, wire, "baseline_id", "baseline-9")
	assertContainsKey(t, wire, "ship_creates")
	assertContainsKey(t, wire, "bullet_updates")
	assertContainsKey(t, wire, "asteroid_updates")
	assertNotContainsKey(t, wire, "bullet_creates")
	assertNotContainsKey(t, wire, "bullet_deletes")
	assertNotContainsKey(t, wire, "asteroid_creates")
	assertNotContainsKey(t, wire, "asteroid_deletes")
	assertContainsKey(t, wire, "pickup_creates")
	assertNotNakedDeltaPayload(t, wire)
}

func TestActiveWirePacketEncodingUsesBulletLifecycleEnvelope(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(BulletWireDeltaPacket{
		Type: PacketFamilyBulletsLifecycle, Metadata: Metadata{Lane: LaneBulletsLifecycle, Sequence: 11, BaselineID: "bullet-baseline-11", SnapshotID: "bullet-snapshot-11", SnapshotKind: SnapshotKind("delta")},
		BulletCreates: []WorldBulletWireRecord{{ID: "bullet-a", OwnerID: "ship-a", X: 1, Y: 2, Rotation: 3, WeaponID: "pulse", ProjectileType: "laser"}},
		BulletDeletes: []string{"bullet-b"},
	}, nil)

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilyBulletsLifecycle)
	assertIntValue(t, wire, "sequence", 11)
	assertStringValue(t, wire, "baseline_id", "bullet-baseline-11")
	assertContainsKey(t, wire, "bullet_creates")
	assertContainsKey(t, wire, "bullet_deletes")
	assertNotContainsKey(t, wire, "bullet_updates")
	assertNotNakedDeltaPayload(t, wire)
}

func TestActiveWirePacketEncodingUsesOverlayDeltaEnvelope(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(OverlayLaneDelta{
		Metadata: Metadata{
			Lane:         LaneOverlay,
			Sequence:     12,
			BaselineID:   "overlay-baseline-12",
			SnapshotID:   "overlay-snapshot-12",
			SnapshotKind: SnapshotKind("delta"),
		},
		Receiver: FieldRecordDelta[OverlayReceiverRecord]{Updates: []map[string]any{{"self_id": "player-1", "score": 10, "primary_cooldown_remaining": 1.25}}},
	}, nil)

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketTypeOverlayDelta)
	assertIntValue(t, wire, "sequence", 12)
	assertIntValue(t, wire, "baseline_sequence", 12)
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
	assertIntValue(t, wire, "sequence", 12)
	assertIntValue(t, wire, "baseline_sequence", 12)

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
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(OverlayWireFullPacket{
		Type:     PacketFamilyOverlayFull,
		Metadata: Metadata{Lane: LaneOverlay, Sequence: 3},
		Receiver: OverlayReceiverWireRecord{
			SelfID:                     "player-1",
			Lives:                      2,
			Score:                      9,
			RespawnCooldown:            1250,
			PrimaryWeaponID:            "pulse",
			PrimaryAmmoPolicy:          "limited",
			PrimaryCooldownRemaining:   500,
			PrimaryAmmoRemaining:       12,
			SecondaryWeaponID:          "mine",
			SecondaryAmmoPolicy:        "infinite",
			SecondaryCooldownRemaining: 750,
			SecondaryAmmoRemaining:     3,
		},
	}, nil))

	assertInt64Value(t, wire, "respawn_cooldown", 1250)
	assertInt64Value(t, wire, "primary_cooldown_remaining", 500)
	assertInt64Value(t, wire, "secondary_cooldown_remaining", 750)
}

func TestWireOverlayWireDeltaPacketEncodesIntegerCooldownUpdates(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(OverlayWireLaneDelta{
		Metadata: Metadata{Lane: LaneOverlay, Sequence: 12, BaselineID: "overlay-baseline-12", SnapshotID: "overlay-snapshot-12", SnapshotKind: SnapshotKind("delta")},
		Receiver: FieldRecordDelta[OverlayReceiverWireRecord]{Updates: []map[string]any{{"self_id": "player-1", "primary_cooldown_remaining": int64(500)}}},
	}, nil))

	updates := mustSliceValue(t, wire, "receiver_updates")
	if len(updates) != 1 {
		t.Fatalf("expected one receiver update, got %#v", updates)
	}
	update := mustMapValue(t, updates[0])
	assertInt64Value(t, update, "primary_cooldown_remaining", 500)
}

func TestActiveWirePacketEncodingUsesSessionDeltaEnvelope(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(SessionLaneDelta{
		Metadata: Metadata{
			Lane:         LaneSession,
			Sequence:     14,
			BaselineID:   "session-baseline-14",
			SnapshotID:   "session-snapshot-14",
			SnapshotKind: SnapshotKind("delta"),
		},
		Players:         FieldRecordDelta[SessionPlayerRecord]{Updates: []map[string]any{{"id": "player-1", "score": 10, "lives": 2}}},
		PlayerLifecycle: FieldRecordDelta[SessionLifecycleRecord]{Updates: []map[string]any{{"player_id": "player-1", "status": "respawning"}}},
		TotalAsteroids:  RecordDelta[SessionTotalAsteroidsRecord]{Updates: []SessionTotalAsteroidsRecord{{ID: "session-14", Count: 8}}},
	}, nil)

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketTypeSessionDelta)
	assertIntValue(t, wire, "sequence", 14)
	assertIntValue(t, wire, "baseline_sequence", 14)
	assertNotContainsKey(t, wire, "players")
	assertContainsKey(t, wire, "player_session_updates")
	assertNotContainsKey(t, wire, "player_lifecycle")
	assertContainsKey(t, wire, "player_lifecycle_updates")
	assertContainsKey(t, wire, "total_asteroids")
	assertNotNakedSessionDeltaPayload(t, wire)
}

func TestActiveWirePacketEncodingUsesLowercaseOverlayShape(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(OverlayFullPacket{
		Type: PacketFamilyOverlayFull,
		Metadata: Metadata{
			Lane:     LaneOverlay,
			Sequence: 3,
		},
		Receiver: OverlayReceiverRecord{
			SelfID:                     "player-1",
			Lives:                      2,
			Score:                      9,
			RespawnCooldown:            1.25,
			PrimaryWeaponID:            "pulse",
			PrimaryAmmoPolicy:          "limited",
			PrimaryCooldownRemaining:   0.5,
			PrimaryAmmoRemaining:       12,
			SecondaryWeaponID:          "mine",
			SecondaryAmmoPolicy:        "infinite",
			SecondaryCooldownRemaining: 0.75,
			SecondaryAmmoRemaining:     3,
		},
	}, nil)

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilyOverlayFull)
	assertStringValue(t, wire, "self_id", "player-1")
	assertContainsKey(t, wire, "respawn_cooldown")
	assertNotContainsKey(t, wire, "respawn")
	assertNotContainsKey(t, wire, "Receiver")
}

func TestActiveWirePacketEncodingUsesLowercaseSessionShape(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(SessionFullPacket{
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
	}, nil)

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilySessionFull)
	assertContainsKey(t, wire, "players")
	assertContainsKey(t, wire, "player_lifecycle")
	assertIntValue(t, wire, "total_asteroids", 42)
}

func TestActiveWirePacketEncodingUsesLowercaseEventShape(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(EventBatchPacket{
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
	}, nil)

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilyEventBatch)
	assertContainsKey(t, wire, "events")
	events := mustSliceValue(t, wire, "events")
	bulletBlast := mustMapValue(t, events[0])
	assertStringValue(t, bulletBlast, "event_id", "event-1")
	assertStringValue(t, bulletBlast, "type", "bullet_blast")
	assertJSONIntValue(t, bulletBlast, "x", 10)
	assertJSONIntValue(t, bulletBlast, "y", 20)

	shipDeath := mustMapValue(t, events[1])
	assertStringValue(t, shipDeath, "event_id", "event-2")
	assertStringValue(t, shipDeath, "type", "ship_death")
	assertStringValue(t, shipDeath, "player_id", "player-1")
	assertIntValue(t, shipDeath, "lives", 2)
	assertJSONIntValue(t, shipDeath, "respawn_delay", 3500)
	assertJSONIntValue(t, shipDeath, "x", 40)
	assertJSONIntValue(t, shipDeath, "y", 50)
}

func TestWireLanePacketRejectsUnsupportedPayloads(t *testing.T) {
	if _, err := WireLanePacket(RealtimeLaneCandidate{Payload: invalidRealtimeLanePayload{}}); err == nil {
		t.Fatal("expected unsupported payload to fail closed")
	}
}

func TestWireWorldFullPacketOmitsInferableRuntimeMetadata(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(WorldFullPacket{
		Type: PacketFamilyWorldFull,
		Metadata: Metadata{
			Lane:           LaneWorld,
			Sequence:       9,
			BaselineID:     "world-baseline-9",
			SnapshotID:     "world-baseline-9",
			SnapshotKind:   SnapshotKind("full"),
			ServerSentMsec: 123,
			ChunkIndex:     0,
			ChunkCount:     1,
			IsFinalChunk:   true,
		},
	}, nil))

	assertStringValue(t, wire, "type", PacketFamilyWorldFull)
	assertIntValue(t, wire, "sequence", 9)
	assertIntValue(t, wire, "server_sent_msec", 123)
	for _, key := range []string{"lane", "baseline_id", "baseline_sequence", "snapshot_id", "snapshot_kind", "chunk_index", "chunk_count", "is_final_chunk"} {
		assertNotContainsKey(t, wire, key)
	}
}

func TestWireWorldDeltaPacketEmitsChunkMetadataOnlyForChunkedPackets(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(WorldDeltaPacket{
		Type: PacketTypeWorldDelta,
		Metadata: Metadata{
			Lane:         LaneWorld,
			Sequence:     10,
			BaselineID:   "world-baseline-9",
			SnapshotID:   "world-snapshot-10",
			SnapshotKind: SnapshotKind("delta"),
			ChunkIndex:   1,
			ChunkCount:   3,
		},
	}, nil))

	assertIntValue(t, wire, "baseline_sequence", 9)
	assertIntValue(t, wire, "chunk_index", 1)
	assertIntValue(t, wire, "chunk_count", 3)
	assertNotContainsKey(t, wire, "is_final_chunk")
}
func TestWireEventBatchPacketOmitsEnvelopeMetadata(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent, Sequence: 11, SnapshotKind: SnapshotKind("delta"), BaselineID: "event-baseline", SnapshotID: "event-snapshot", ChunkIndex: 1, ChunkCount: 3, IsFinalChunk: true},
		Batch:    EventBatchRecord{BatchID: "event-batch-11"},
	}, nil))

	assertStringValue(t, wire, "type", PacketFamilyEventBatch)
	assertIntValue(t, wire, "sequence", 11)
	assertNotContainsKey(t, wire, "lane")
	assertNotContainsKey(t, wire, "baseline_id")
	assertNotContainsKey(t, wire, "snapshot_id")
	assertNotContainsKey(t, wire, "snapshot_kind")
	assertNotContainsKey(t, wire, "chunk_index")
	assertNotContainsKey(t, wire, "chunk_count")
	assertNotContainsKey(t, wire, "is_final_chunk")
}
func TestWireEventBatchPacketShapesBulletBlastWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent},
		Batch: EventBatchRecord{
			Events: []EventRecord{{
				EventID: "event-1",
				Event: game.EventState{
					Type:       "bullet_blast",
					X:          12.34,
					Y:          56.78,
					PickupID:   "pickup-1",
					PickupType: "shield",
					TableID:    "table-1",
					EffectType: "blast",
					Amount:     9,
					LivesAfter: 1,
					SourceID:   "ship-1",
					SourceType: "ship",
				},
			}},
		},
	}, nil))

	events := mustSliceValue(t, wire, "events")
	record := mustMapValue(t, events[0])
	assertStringValue(t, record, "event_id", "event-1")
	assertStringValue(t, record, "type", "bullet_blast")
	assertInt64Value(t, record, "x", 123)
	assertInt64Value(t, record, "y", 568)
	for _, key := range []string{"pickup_id", "pickup_type", "table_id", "effect_type", "amount", "lives_after", "source_id", "source_type", "player_id", "lives", "respawn_delay"} {
		assertNotContainsKey(t, record, key)
	}
}

func TestWireEventBatchPacketShapesShipDeathWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent},
		Batch: EventBatchRecord{
			Events: []EventRecord{{
				EventID: "event-2",
				Event: game.EventState{
					Type:         "ship_death",
					PlayerID:     "player-1",
					Lives:        2,
					RespawnDelay: 3.5,
					X:            1979.580796080448,
					Y:            235.79718289389993,
					PickupID:     "pickup-1",
					PickupType:   "shield",
					TableID:      "table-1",
					EffectType:   "death",
					Amount:       7,
					LivesAfter:   1,
					SourceID:     "ship-2",
					SourceType:   "ship",
				},
			}},
		},
	}, nil))

	events := mustSliceValue(t, wire, "events")
	record := mustMapValue(t, events[0])
	assertStringValue(t, record, "event_id", "event-2")
	assertStringValue(t, record, "type", "ship_death")
	assertStringValue(t, record, "player_id", "player-1")
	assertIntValue(t, record, "lives", 2)
	assertInt64Value(t, record, "respawn_delay", 3500)
	assertInt64Value(t, record, "x", 19796)
	assertInt64Value(t, record, "y", 2358)
	for _, key := range []string{"pickup_id", "pickup_type", "table_id", "effect_type", "amount", "lives_after", "source_id", "source_type"} {
		assertNotContainsKey(t, record, key)
	}
}
func TestWireEventBatchPacketEncodesHighPrecisionShipDeathFloatsAsIntegers(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent},
		Batch: EventBatchRecord{
			Events: []EventRecord{{
				EventID: "event-precision-ship",
				Event: game.EventState{
					Type:         "ship_death",
					PlayerID:     "player-1",
					Lives:        2,
					RespawnDelay: 3,
					X:            1979.580796080448,
					Y:            235.79718289389993,
				},
			}},
		},
	}, nil)

	encoded := mustEncodeWirePacket(t, candidate)
	encodedString := string(encoded)
	for _, fragment := range []string{"1979.580796080448", "235.79718289389993"} {
		if strings.Contains(encodedString, fragment) {
			t.Fatalf("expected encoded JSON to omit raw float %q, got %s", fragment, encodedString)
		}
	}

	record := mustMapValue(t, mustSliceValue(t, mustDecodeWirePacket(t, encoded), "events")[0])
	assertJSONIntValue(t, record, "respawn_delay", 3000)
	assertJSONIntValue(t, record, "x", 19796)
	assertJSONIntValue(t, record, "y", 2358)
}

func TestWireEventBatchPacketKeepsLegacyFallbackForUnknownEventTypes(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent},
		Batch: EventBatchRecord{
			Events: []EventRecord{{
				EventID: "event-3",
				Event: game.EventState{
					Type:       "unknown_event",
					PlayerID:   "player-1",
					PickupID:   "pickup-1",
					PickupType: "shield",
					X:          1.25,
					Y:          2.5,
				},
			}},
		},
	}, nil))

	events := mustSliceValue(t, wire, "events")
	record := mustMapValue(t, events[0])
	assertStringValue(t, record, "event_id", "event-3")
	assertStringValue(t, record, "type", "unknown_event")
	assertStringValue(t, record, "player_id", "player-1")
	assertStringValue(t, record, "pickup_id", "pickup-1")
	assertStringValue(t, record, "pickup_type", "shield")
	assertFloatValue(t, record, "x", 1.25)
	assertFloatValue(t, record, "y", 2.5)
}
func TestWireEventBatchPacketShapesDamageAppliedWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent},
		Batch: EventBatchRecord{
			Events: []EventRecord{{
				EventID: "event-damage-applied",
				Event: game.EventState{
					Type:         "damage_applied",
					SourceType:   "projectile",
					SourceID:     "bullet-1",
					EffectType:   "explosive",
					Amount:       17,
					X:            12.34,
					Y:            56.78,
					PlayerID:     "player-1",
					Lives:        2,
					RespawnDelay: 3.5,
					PickupID:     "pickup-1",
					PickupType:   "shield",
					TableID:      "table-1",
					LivesAfter:   1,
				},
			}},
		},
	}, nil))

	record := mustMapValue(t, mustSliceValue(t, wire, "events")[0])
	assertStringValue(t, record, "event_id", "event-damage-applied")
	assertStringValue(t, record, "type", "damage_applied")
	assertStringValue(t, record, "source_type", "projectile")
	assertStringValue(t, record, "source_id", "bullet-1")
	assertStringValue(t, record, "effect_type", "explosive")
	assertIntValue(t, record, "amount", 17)
	assertInt64Value(t, record, "x", 123)
	assertInt64Value(t, record, "y", 568)
	for _, key := range []string{"player_id", "lives", "respawn_delay", "pickup_id", "pickup_type", "table_id", "lives_after"} {
		assertNotContainsKey(t, record, key)
	}
}
func TestWireEventBatchPacketEncodesHighPrecisionDamageAppliedFloatsAsIntegers(t *testing.T) {
	candidate := mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent},
		Batch: EventBatchRecord{
			Events: []EventRecord{{
				EventID: "event-precision-damage",
				Event: game.EventState{
					Type:       "damage_applied",
					SourceType: "projectile",
					SourceID:   "bullet-1",
					X:          12.3456789012345,
					Y:          98.7654321098765,
				},
			}},
		},
	}, nil)

	encoded := mustEncodeWirePacket(t, candidate)
	encodedString := string(encoded)
	for _, fragment := range []string{"12.3456789012345", "98.7654321098765"} {
		if strings.Contains(encodedString, fragment) {
			t.Fatalf("expected encoded JSON to omit raw float %q, got %s", fragment, encodedString)
		}
	}

	record := mustMapValue(t, mustSliceValue(t, mustDecodeWirePacket(t, encoded), "events")[0])
	assertJSONIntValue(t, record, "x", 123)
	assertJSONIntValue(t, record, "y", 988)
}

func TestWireEventBatchPacketShapesDamageOverTimeStartedWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent},
		Batch: EventBatchRecord{
			Events: []EventRecord{{
				EventID: "event-dot-started",
				Event: game.EventState{
					Type:       "damage_over_time_started",
					SourceType: "asteroid",
					SourceID:   "hazard-1",
					EffectType: "radioactive",
					Amount:     2,
					X:          9.99,
					Y:          8.88,
					PlayerID:   "player-1",
					PickupID:   "pickup-1",
					TableID:    "table-1",
				},
			}},
		},
	}, nil))

	record := mustMapValue(t, mustSliceValue(t, wire, "events")[0])
	assertStringValue(t, record, "event_id", "event-dot-started")
	assertStringValue(t, record, "type", "damage_over_time_started")
	assertStringValue(t, record, "source_type", "asteroid")
	assertStringValue(t, record, "source_id", "hazard-1")
	assertStringValue(t, record, "effect_type", "radioactive")
	assertIntValue(t, record, "amount", 2)
	for _, key := range []string{"x", "y", "player_id", "pickup_id", "pickup_type", "table_id", "lives", "respawn_delay"} {
		assertNotContainsKey(t, record, key)
	}
}

func TestWireEventBatchPacketShapesDamageOverTimeTickWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{
		Type:     PacketFamilyEventBatch,
		Metadata: Metadata{Lane: LaneEvent},
		Batch: EventBatchRecord{
			Events: []EventRecord{{
				EventID: "event-dot-tick",
				Event: game.EventState{
					Type:       "damage_over_time_tick",
					SourceType: "asteroid",
					SourceID:   "hazard-1",
					EffectType: "radioactive",
					Amount:     3,
					X:          45.67,
					Y:          89.01,
					PlayerID:   "player-1",
					Lives:      2,
					PickupID:   "pickup-1",
					TableID:    "table-1",
				},
			}},
		},
	}, nil))

	record := mustMapValue(t, mustSliceValue(t, wire, "events")[0])
	assertStringValue(t, record, "event_id", "event-dot-tick")
	assertStringValue(t, record, "type", "damage_over_time_tick")
	assertStringValue(t, record, "source_type", "asteroid")
	assertStringValue(t, record, "source_id", "hazard-1")
	assertStringValue(t, record, "effect_type", "radioactive")
	assertIntValue(t, record, "amount", 3)
	assertInt64Value(t, record, "x", 457)
	assertInt64Value(t, record, "y", 890)
	for _, key := range []string{"player_id", "lives", "respawn_delay", "pickup_id", "pickup_type", "table_id", "lives_after"} {
		assertNotContainsKey(t, record, key)
	}
}

func TestWireEventBatchPacketShapesRadialEffectStartedWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: Metadata{Lane: LaneEvent}, Batch: EventBatchRecord{Events: []EventRecord{{
		EventID: "event-radial",
		Event:   game.EventState{Type: "radial_effect_started", SourceType: "pickup", SourceID: "pickup-1", EffectType: "pulse", Amount: 9, X: 10.25, Y: 20.5, PlayerID: "player-1", PickupID: "pickup-1", TableID: "table-1"},
	}}}}, nil))

	record := mustMapValue(t, mustSliceValue(t, wire, "events")[0])
	assertStringValue(t, record, "event_id", "event-radial")
	assertStringValue(t, record, "type", "radial_effect_started")
	assertStringValue(t, record, "source_type", "pickup")
	assertStringValue(t, record, "source_id", "pickup-1")
	assertStringValue(t, record, "effect_type", "pulse")
	assertInt64Value(t, record, "x", 103)
	assertInt64Value(t, record, "y", 205)
	for _, key := range []string{"amount", "player_id", "pickup_id", "pickup_type", "table_id", "lives", "respawn_delay"} {
		assertNotContainsKey(t, record, key)
	}
}

func TestWireEventBatchPacketShapesPickupCollectedWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: Metadata{Lane: LaneEvent}, Batch: EventBatchRecord{Events: []EventRecord{{
		EventID: "event-pickup-collected",
		Event:   game.EventState{Type: "pickup_collected", PlayerID: "player-1", PickupID: "pickup-1", PickupType: "shield", X: 12.5, Y: 34.5, SourceType: "ship", SourceID: "ship-1", TableID: "table-1", Lives: 2},
	}}}}, nil))

	record := mustMapValue(t, mustSliceValue(t, wire, "events")[0])
	assertStringValue(t, record, "event_id", "event-pickup-collected")
	assertStringValue(t, record, "type", "pickup_collected")
	assertStringValue(t, record, "player_id", "player-1")
	assertStringValue(t, record, "pickup_id", "pickup-1")
	assertStringValue(t, record, "pickup_type", "shield")
	assertInt64Value(t, record, "x", 125)
	assertInt64Value(t, record, "y", 345)
	for _, key := range []string{"source_type", "source_id", "table_id", "lives", "respawn_delay", "effect_type", "amount"} {
		assertNotContainsKey(t, record, key)
	}
}

func TestWireEventBatchPacketShapesPickupEffectAppliedWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: Metadata{Lane: LaneEvent}, Batch: EventBatchRecord{Events: []EventRecord{{
		EventID: "event-pickup-effect",
		Event:   game.EventState{Type: "pickup_effect_applied", PlayerID: "player-1", PickupID: "pickup-1", PickupType: "shield", EffectType: "repair", Amount: 4, LivesAfter: 3, X: 1.5, Y: 2.5, TableID: "table-1"},
	}}}}, nil))

	record := mustMapValue(t, mustSliceValue(t, wire, "events")[0])
	assertStringValue(t, record, "event_id", "event-pickup-effect")
	assertStringValue(t, record, "type", "pickup_effect_applied")
	assertStringValue(t, record, "player_id", "player-1")
	assertStringValue(t, record, "pickup_id", "pickup-1")
	assertStringValue(t, record, "pickup_type", "shield")
	assertStringValue(t, record, "effect_type", "repair")
	assertIntValue(t, record, "amount", 4)
	assertIntValue(t, record, "lives_after", 3)
	for _, key := range []string{"x", "y", "source_type", "source_id", "table_id", "respawn_delay"} {
		assertNotContainsKey(t, record, key)
	}
}

func TestWireEventBatchPacketShapesPickupExpiredWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: Metadata{Lane: LaneEvent}, Batch: EventBatchRecord{Events: []EventRecord{{
		EventID: "event-pickup-expired",
		Event:   game.EventState{Type: "pickup_expired", PickupID: "pickup-1", PickupType: "shield", X: 22.2, Y: 33.3, PlayerID: "player-1", SourceID: "ship-1"},
	}}}}, nil))

	record := mustMapValue(t, mustSliceValue(t, wire, "events")[0])
	assertStringValue(t, record, "event_id", "event-pickup-expired")
	assertStringValue(t, record, "type", "pickup_expired")
	assertStringValue(t, record, "pickup_id", "pickup-1")
	assertStringValue(t, record, "pickup_type", "shield")
	assertInt64Value(t, record, "x", 222)
	assertInt64Value(t, record, "y", 333)
	for _, key := range []string{"player_id", "source_type", "source_id", "table_id", "effect_type", "amount", "lives"} {
		assertNotContainsKey(t, record, key)
	}
}

func TestWireEventBatchPacketShapesPickupDroppedWithRelevantFieldsOnly(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(EventBatchPacket{Type: PacketFamilyEventBatch, Metadata: Metadata{Lane: LaneEvent}, Batch: EventBatchRecord{Events: []EventRecord{{
		EventID: "event-pickup-dropped",
		Event:   game.EventState{Type: "pickup_dropped", PickupID: "pickup-1", PickupType: "shield", SourceType: "ship", SourceID: "ship-1", TableID: "table-1", X: 44.4, Y: 55.5, PlayerID: "player-1", Lives: 2},
	}}}}, nil))

	record := mustMapValue(t, mustSliceValue(t, wire, "events")[0])
	assertStringValue(t, record, "event_id", "event-pickup-dropped")
	assertStringValue(t, record, "type", "pickup_dropped")
	assertStringValue(t, record, "pickup_id", "pickup-1")
	assertStringValue(t, record, "pickup_type", "shield")
	assertStringValue(t, record, "source_type", "ship")
	assertStringValue(t, record, "source_id", "ship-1")
	assertStringValue(t, record, "table_id", "table-1")
	assertInt64Value(t, record, "x", 444)
	assertInt64Value(t, record, "y", 555)
	for _, key := range []string{"player_id", "lives", "respawn_delay", "effect_type", "amount", "lives_after"} {
		assertNotContainsKey(t, record, key)
	}
}

func mustEncodeWirePacket(t *testing.T, candidate RealtimeLaneCandidate) []byte {
	t.Helper()

	encoded, err := packetcodec.Encode(mustWireLanePacket(t, candidate))
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
	candidate := mustRealtimeLaneCandidate(WorldFullPacket{
		Type:      PacketFamilyWorldFull,
		Metadata:  Metadata{Lane: LaneWorld, Sequence: 21},
		Ships:     []WorldShipRecord{{ID: "ship-1", ShipType: "v_wing", X: 1, Y: 2, Rotation: 3, Health: 4, Shields: 5, Thrusting: true, TargetKind: "player", TargetID: "player-1"}},
		Bullets:   []WorldBulletRecord{{ID: "bullet-1", OwnerID: "ship-1", X: 6, Y: 7, Rotation: 8, WeaponID: "pulse", ProjectileType: "laser"}},
		Asteroids: []WorldAsteroidRecord{{ID: "asteroid-1", X: 9, Y: 10, Size: 2, Health: 11, Scale: 1.5, Variant: 3}},
		Pickups:   []WorldPickupRecord{{ID: "pickup-1", Type: "shield", PickupClass: "armor", X: 12, Y: 13, Health: 1, AgeSeconds: 4.5, LifespanSeconds: 9.5}},
	}, nil)

	wire := mustDecodeWirePacket(t, mustEncodeWirePacket(t, candidate))

	assertStringValue(t, wire, "type", PacketFamilyWorldFull)
	assertContainsKey(t, wire, "ships")
	assertContainsKey(t, wire, "bullets")
	assertContainsKey(t, wire, "asteroids")
	assertContainsKey(t, wire, "pickups")
}

func TestWireSessionWireFullPacketEncodesIntegerCooldownFields(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(SessionWireFullPacket{
		Type:     PacketFamilySessionFull,
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
		TotalAsteroids:  42,
	}, nil))

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
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(WorldFullPacket{Type: PacketFamilyWorldFull, Metadata: Metadata{Lane: LaneWorld}}, nil))

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
	fullMetadata, ok := RealtimeLaneCandidate{Payload: full}.Metadata()
	if !ok {
		t.Fatal("expected world wire full metadata to be found")
	}
	if fullMetadata != full.Metadata {
		t.Fatalf("full metadata = %#v, want %#v", fullMetadata, full.Metadata)
	}

	delta := WorldWireDeltaPacket{Type: PacketTypeWorldDelta, Metadata: Metadata{Lane: LaneWorld, Sequence: 22, SnapshotKind: SnapshotKind("delta")}}
	deltaMetadata, ok := mustRealtimeLaneCandidate(delta, nil).Metadata()
	if !ok {
		t.Fatal("expected world wire delta metadata to be found")
	}
	if deltaMetadata != delta.Metadata {
		t.Fatalf("delta metadata = %#v, want %#v", deltaMetadata, delta.Metadata)
	}
}

func TestWireWorldWireFullPacketEncodesIntegerWorldFields(t *testing.T) {
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(WorldWireFullPacket{
		Type:      PacketFamilyWorldFull,
		Metadata:  Metadata{Lane: LaneWorld, Sequence: 7},
		Ships:     []WorldShipWireRecord{{ID: "ship-1", ShipType: "v_wing", X: 10, Y: 20, Rotation: 30, Health: 4, Shields: 5, Thrusting: true, TargetKind: "player", TargetID: "player-1"}},
		Bullets:   []WorldBulletWireRecord{{ID: "bullet-1", OwnerID: "ship-1", X: 6, Y: 7, Rotation: 8, WeaponID: "pulse", ProjectileType: "laser"}},
		Asteroids: []WorldAsteroidWireRecord{{ID: "asteroid-1", X: 9, Y: 10, Size: 2, Health: 11, Scale: 15, Variant: 3}},
		Pickups:   []WorldPickupWireRecord{{ID: "pickup-1", Type: "shield", PickupClass: "armor", X: 12, Y: 13, Health: 1, AgeSeconds: 4, LifespanSeconds: 9}},
	}, nil))

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
	wire := mustWireLanePacket(t, mustRealtimeLaneCandidate(WorldWireDeltaPacket{
		Type:     PacketTypeWorldDelta,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 9, SnapshotKind: SnapshotKind("delta")},
		Ships:    FieldRecordDelta[WorldShipWireRecord]{Updates: []map[string]any{{"id": "ship-1", "x": int64(10), "y": int64(20), "rotation": int64(30), "thrusting": true}}},
	}, nil))

	updates := mustSliceValue(t, wire, "ship_updates")
	update := mustMapValue(t, updates[0])
	assertInt64Value(t, update, "x", 10)
	assertInt64Value(t, update, "y", 20)
	assertInt64Value(t, update, "rotation", 30)
	assertNotContainsKey(t, update, "ship_type")
}
