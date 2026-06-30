package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestCompareShadowRealtimeOwnershipAndMetadata(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID:         "player-1",
		Lives:          3,
		ServerSentMsec: 1234,
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing", X: 1, Y: 2, Rotation: 3, Health: 4, Shields: 5, Thrusting: true, TargetKind: "player", TargetID: "player-2", PrimaryCooldownRemaining: 99, PrimaryAmmoRemaining: 88, SecondaryCooldownRemaining: 77, SecondaryAmmoRemaining: 66},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 9, Lives: 3, RespawnCooldown: 1.5, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite", SecondaryWeaponID: "bomb", SecondaryAmmoPolicy: "limited", SpawnX: 7, SpawnY: 8},
		},
		PlayerLifecycle: map[string]string{"player-1": "active"},
		Bullets: map[string]runtime.BulletState{
			"bullet-1": {ID: "bullet-1", OwnerID: "player-1", X: 9, Y: 10, Rotation: 11, WeaponID: "laser", ProjectileType: "bolt"},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-1": {ID: "asteroid-1", X: 12, Y: 13, Size: 1, Health: 2, Scale: 3, Variant: 4},
		},
		Pickups: map[string]runtime.PickupState{
			"pickup-1": {ID: "pickup-1", Type: "ammo", PickupClass: "weapon", X: 14, Y: 15, Health: 16, AgeSeconds: 17, LifespanSeconds: 18},
		},
		TotalAsteroids: 19,
		PendingEvents: []game.PendingPresentationEvent{{EventID: "event-1", Event: game.EventState{Type: "ship_death"}}},
	}

	world := BuildWorldFullPacket(snapshot, 7)
	overlay := BuildOverlayFullPacket(snapshot, "player-1", 8)
	session := BuildSessionFullPacket(snapshot, 9)
	events := BuildEventBatchPacket(snapshot.PendingEvents, 10, snapshot.ServerSentMsec)

	if world.Metadata.Lane != LaneWorld || overlay.Metadata.Lane != LaneOverlay || session.Metadata.Lane != LaneSession || events.Metadata.Lane != LaneEvent {
		t.Fatalf("expected packet lanes to match ownership, got world=%q overlay=%q session=%q events=%q", world.Metadata.Lane, overlay.Metadata.Lane, session.Metadata.Lane, events.Metadata.Lane)
	}
	if world.Metadata.ServerSentMsec != snapshot.ServerSentMsec || overlay.Metadata.ServerSentMsec != snapshot.ServerSentMsec || session.Metadata.ServerSentMsec != snapshot.ServerSentMsec || events.Metadata.ServerSentMsec != snapshot.ServerSentMsec {
		t.Fatalf("expected server_sent_msec metadata to be preserved, got world=%d overlay=%d session=%d events=%d", world.Metadata.ServerSentMsec, overlay.Metadata.ServerSentMsec, session.Metadata.ServerSentMsec, events.Metadata.ServerSentMsec)
	}
	if got := len(world.Ships); got != 1 || world.Ships[0].ID != "player-1" || world.Ships[0].Health != 4 || world.Ships[0].TargetKind != "player" || world.Ships[0].TargetID != "player-2" {
		t.Fatalf("world lane ownership mismatch: %#v", world.Ships)
	}
	if got := overlay.Receiver; got.SelfID != "player-1" || got.Lives != 3 || got.Score != 9 {
		t.Fatalf("overlay lane ownership mismatch: %#v", overlay.Receiver)
	}
	if got := len(session.Players); got != 1 || session.Players[0].ID != "player-1" || session.Players[0].Score != 9 || session.PlayerLifecycle[0].PlayerID != "player-1" {
		t.Fatalf("session lane ownership mismatch: players=%#v lifecycle=%#v", session.Players, session.PlayerLifecycle)
	}
	if got := len(events.Batch.Events); got != 1 || events.Batch.Events[0].EventID != "event-1" || events.Batch.Events[0].Event.Type != "ship_death" {
		t.Fatalf("event_batch ownership mismatch: %#v", events.Batch.Events)
	}
}
