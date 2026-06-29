package realtime

import (
	"encoding/json"
	"strings"
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
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

func assertIntValue(t *testing.T, wire map[string]any, key string, want int) {
	t.Helper()

	value, ok := wire[key]
	if !ok {
		t.Fatalf("expected key %q to exist", key)
	}
	got, ok := value.(float64)
	if !ok {
		t.Fatalf("expected key %q to be numeric, got %#v", key, value)
	}
	if int(got) != want {
		t.Fatalf("key %q = %v, want %d", key, got, want)
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

