package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestProjectWorldLaneFieldOwnershipAndOrder(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		Players: map[string]runtime.ShipState{
			"ship-b": {
				ID:         "ship-b",
				ShipType:   "v_wing",
				X:          20,
				Y:          30,
				Rotation:   0.2,
				Health:     2,
				Shields:    1,
				Thrusting:  true,
				TargetKind: "player",
				TargetID:   "p-2",
				// world lane should not include these local overlay fields
				PrimaryCooldownRemaining: 99,
				PrimaryAmmoRemaining:     88,
				SecondaryCooldownRemaining: 77,
				SecondaryAmmoRemaining:   66,
			},
			"ship-a": {
				ID:         "ship-a",
				ShipType:   "v_wing",
				X:          10,
				Y:          15,
				Rotation:   0.1,
				Health:     3,
				Shields:    2,
				Thrusting:  false,
				TargetKind: "player",
				TargetID:   "p-1",
				PrimaryCooldownRemaining: 55,
				PrimaryAmmoRemaining:     44,
				SecondaryCooldownRemaining: 33,
				SecondaryAmmoRemaining:   22,
			},
		},
		Bullets: map[string]runtime.BulletState{
			"bullet-b": {ID: "bullet-b", OwnerID: "ship-b", X: 5, Y: 6, Rotation: 0.6, WeaponID: "basic", ProjectileType: "cannon"},
			"bullet-a": {ID: "bullet-a", OwnerID: "ship-a", X: 3, Y: 4, Rotation: 0.4, WeaponID: "basic", ProjectileType: "cannon"},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-b": {ID: "asteroid-b", X: 50, Y: 60, Size: 3, Health: 4, Scale: 1.5, Variant: 2},
			"asteroid-a": {ID: "asteroid-a", X: 40, Y: 45, Size: 2, Health: 5, Scale: 1.0, Variant: 1},
		},
		Pickups: map[string]runtime.PickupState{
			"pickup-b": {ID: "pickup-b", Type: "1_up", PickupClass: "life", X: 7, Y: 8, Health: 1, AgeSeconds: 2, LifespanSeconds: 10},
			"pickup-a": {ID: "pickup-a", Type: "shield", PickupClass: "armor", X: 1, Y: 2, Health: 1, AgeSeconds: 1, LifespanSeconds: 5},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"ship-a": {ID: "ship-a", ShipType: "v_wing", Score: 9, Lives: 3, RespawnCooldown: 0.5, PrimaryWeaponID: "basic", PrimaryAmmoPolicy: "infinite", SecondaryWeaponID: "mine", SecondaryAmmoPolicy: "limited", SpawnX: 1, SpawnY: 2},
		},
		PlayerLifecycle: map[string]string{"ship-a": "active"},
		PendingEvents: []game.PendingPresentationEvent{{EventID: "evt-1", Event: game.EventState{Type: "bullet_blast"}}},
	}

	projection := ProjectWorldLane(snapshot)

	if len(projection.Ships) != 2 {
		t.Fatalf("expected 2 ships, got %d", len(projection.Ships))
	}
	if projection.Ships[0].ID != "ship-a" || projection.Ships[1].ID != "ship-b" {
		t.Fatalf("expected ships sorted by ID, got %#v", projection.Ships)
	}
	ship := projection.Ships[0]
	if ship.ShipType != "v_wing" || ship.X != 10 || ship.Y != 15 || ship.Rotation != 0.1 || ship.Health != 3 || ship.Shields != 2 || ship.Thrusting != false || ship.TargetKind != "player" || ship.TargetID != "p-1" {
		t.Fatalf("expected world ship fields to be preserved, got %#v", ship)
	}
	if ship.ShipType == "" || ship.TargetKind == "" || ship.TargetID == "" {
		t.Fatalf("expected ship identity fields to be populated, got %#v", ship)
	}

	if len(projection.Bullets) != 2 || projection.Bullets[0].ID != "bullet-a" || projection.Bullets[1].ID != "bullet-b" {
		t.Fatalf("expected bullets sorted by ID, got %#v", projection.Bullets)
	}
	if len(projection.Asteroids) != 2 || projection.Asteroids[0].ID != "asteroid-a" || projection.Asteroids[1].ID != "asteroid-b" {
		t.Fatalf("expected asteroids sorted by ID, got %#v", projection.Asteroids)
	}
	if len(projection.Pickups) != 2 || projection.Pickups[0].ID != "pickup-a" || projection.Pickups[1].ID != "pickup-b" {
		t.Fatalf("expected pickups sorted by ID, got %#v", projection.Pickups)
	}
}

func TestBuildWorldFullPacketUsesMetadataAndSortedProjection(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:         "snapshot-1",
		ServerSentMsec: 123,
		Players: map[string]runtime.ShipState{
			"ship-b": {ID: "ship-b", ShipType: "v_wing"},
			"ship-a": {ID: "ship-a", ShipType: "v_wing"},
		},
	}

	packet := BuildWorldFullPacket(snapshot, 9)

	if packet.Type != PacketFamilyWorldFull {
		t.Fatalf("expected world full packet type, got %q", packet.Type)
	}
	if packet.Metadata.Lane != LaneWorld || packet.Metadata.Sequence != 9 || packet.Metadata.BaselineID != "snapshot-1" || packet.Metadata.SnapshotID != "snapshot-1" || packet.Metadata.ServerSentMsec != 123 || packet.Metadata.SnapshotKind != SnapshotKind("full") || packet.Metadata.ChunkIndex != 0 || packet.Metadata.ChunkCount != 1 || !packet.Metadata.IsFinalChunk {
		t.Fatalf("expected metadata to be populated, got %#v", packet.Metadata)
	}
	if len(packet.Ships) != 2 || packet.Ships[0].ID != "ship-a" || packet.Ships[1].ID != "ship-b" {
		t.Fatalf("expected ships sorted by ID in packet, got %#v", packet.Ships)
	}
}

