package realtime

import (
	"reflect"
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

var _ func(game.GameplayPresentationSnapshot, RealtimeSessionState) RealtimeLanePlan = AssembleRealtimeLaneCandidates

func TestAssembleRealtimeLaneCandidatesChoosesFullAndDeltaWithoutDraining(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {
				ID:        "player-1",
				ShipType:  "v_wing",
				X:         10,
				Y:         20,
				Rotation:  30,
				Health:    5,
				Shields:   2,
				Thrusting: true,
			},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {
				ID:                "player-1",
				ShipType:          "v_wing",
				Score:             99,
				Lives:             3,
				RespawnCooldown:   1.5,
				PrimaryWeaponID:   "laser",
				PrimaryAmmoPolicy: "infinite",
				SpawnX:            1,
				SpawnY:            2,
			},
		},
		PendingEvents:  []game.PendingPresentationEvent{{EventID: "event-1", Event: game.EventState{Type: "ship_death"}}},
		ServerSentMsec: 1234,
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 2, BaselineID: "world-baseline", IsFinalChunk: true})
	state.UpdateLane(LaneSession, Metadata{Sequence: 3, BaselineID: "session-baseline", IsFinalChunk: true})
	state.UpdateLane(LaneEvent, Metadata{Sequence: 9, IsFinalChunk: true})

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if got, want := len(plan.Candidates), 4; got != want {
		t.Fatalf("candidate count = %d, want %d", got, want)
	}

	if got, want := plan.Candidates[0].Lane, Lane(LaneWorld); got != want {
		t.Fatalf("world lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[0].Delta.(WorldLaneProjection); !ok {
		t.Fatalf("world candidate delta type = %T, want WorldLaneProjection", plan.Candidates[0].Delta)
	} else if len(got.Ships) != 1 || got.Ships[0].ID != "player-1" {
		t.Fatalf("world delta = %#v, want player-1 ship", got)
	}

	if got, want := plan.Candidates[1].Lane, Lane(LaneOverlay); got != want {
		t.Fatalf("overlay lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[1].Full.(OverlayFullPacket); !ok {
		t.Fatalf("overlay candidate full type = %T, want OverlayFullPacket", plan.Candidates[1].Full)
	} else if got.Metadata.Lane != LaneOverlay || got.Receiver.SelfID != "player-1" {
		t.Fatalf("overlay packet = %#v, want player-1 overlay packet", got)
	}

	if got, want := plan.Candidates[2].Lane, Lane(LaneSession); got != want {
		t.Fatalf("session lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[2].Delta.(SessionLaneProjection); !ok {
		t.Fatalf("session candidate delta type = %T, want SessionLaneProjection", plan.Candidates[2].Delta)
	} else if len(got.Players) != 1 || got.Players[0].ID != "player-1" {
		t.Fatalf("session delta = %#v, want player-1 session", got)
	}

	if got, want := plan.Candidates[3].Lane, Lane(LaneEvent); got != want {
		t.Fatalf("event lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[3].Full.(EventBatchPacket); !ok {
		t.Fatalf("event candidate full type = %T, want EventBatchPacket", plan.Candidates[3].Full)
	} else if got.Metadata.Lane != LaneEvent || len(got.Batch.Events) != 1 || got.Batch.Events[0].EventID != "event-1" {
		t.Fatalf("event batch = %#v, want preserved event id", got)
	}

	if len(snapshot.PendingEvents) != 1 || snapshot.PendingEvents[0].EventID != "event-1" {
		t.Fatalf("planner mutated pending events: %#v", snapshot.PendingEvents)
	}
}

func TestRealtimeOwnershipParityAcrossLanes(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Lives: 4,
		Players: map[string]runtime.ShipState{
			"player-1": {
				ID:                         "player-1",
				ShipType:                   "v_wing",
				X:                          10,
				Y:                          20,
				Rotation:                   30,
				Health:                     5,
				Shields:                    2,
				Thrusting:                  true,
				TargetKind:                 "player",
				TargetID:                   "player-2",
				PrimaryWeaponID:            "laser",
				PrimaryAmmoPolicy:          "infinite",
				PrimaryCooldownRemaining:   1.25,
				PrimaryAmmoRemaining:       9,
				SecondaryWeaponID:          "bomb",
				SecondaryAmmoPolicy:        "limited",
				SecondaryCooldownRemaining: 2.5,
				SecondaryAmmoRemaining:     3,
			},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {
				ID:                "player-1",
				ShipType:          "v_wing",
				Score:             99,
				Lives:             4,
				RespawnCooldown:   1.5,
				PrimaryWeaponID:   "laser",
				PrimaryAmmoPolicy: "infinite",
				SecondaryWeaponID: "bomb",
				SecondaryAmmoPolicy: "limited",
				SpawnX:            1,
				SpawnY:            2,
			},
		},
		PlayerLifecycle: map[string]string{"player-1": "active"},
		Bullets: map[string]runtime.BulletState{
			"bullet-1": {ID: "bullet-1", OwnerID: "player-1", X: 1, Y: 2, Rotation: 3, WeaponID: "laser", ProjectileType: "bolt"},
		},
		Asteroids: map[string]runtime.AsteroidState{
			"asteroid-1": {ID: "asteroid-1", X: 4, Y: 5, Size: 6, Health: 7, Scale: 8, Variant: 9},
		},
		Pickups: map[string]runtime.PickupState{
			"pickup-1": {ID: "pickup-1", Type: "ammo", PickupClass: "weapon", X: 6, Y: 7, Health: 8, AgeSeconds: 9, LifespanSeconds: 10},
		},
		TotalAsteroids: 11,
		PendingEvents:  []game.PendingPresentationEvent{{EventID: "event-1", Event: game.EventState{Type: "ship_death"}}},
		ServerSentMsec: 1234,
	}

	world := ProjectWorldLane(snapshot)
	if len(world.Ships) != 1 || world.Ships[0].ShipType != "v_wing" {
		t.Fatalf("world ships = %#v, want ship ownership only", world.Ships)
	}
	worldShipType := reflect.TypeOf(WorldShipRecord{})
	if _, ok := worldShipType.FieldByName("PrimaryWeaponID"); ok {
		t.Fatalf("world ship leaked PrimaryWeaponID field")
	}
	if _, ok := worldShipType.FieldByName("PrimaryAmmoRemaining"); ok {
		t.Fatalf("world ship leaked PrimaryAmmoRemaining field")
	}
	if _, ok := worldShipType.FieldByName("PrimaryCooldownRemaining"); ok {
		t.Fatalf("world ship leaked PrimaryCooldownRemaining field")
	}
	if _, ok := worldShipType.FieldByName("PrimaryAmmoPolicy"); ok {
		t.Fatalf("world ship leaked PrimaryAmmoPolicy field")
	}
	if _, ok := worldShipType.FieldByName("SecondaryWeaponID"); ok {
		t.Fatalf("world ship leaked SecondaryWeaponID field")
	}
	if _, ok := worldShipType.FieldByName("SecondaryCooldownRemaining"); ok {
		t.Fatalf("world ship leaked SecondaryCooldownRemaining field")
	}
	if _, ok := worldShipType.FieldByName("SecondaryAmmoRemaining"); ok {
		t.Fatalf("world ship leaked SecondaryAmmoRemaining field")
	}
	if _, ok := worldShipType.FieldByName("SecondaryAmmoPolicy"); ok {
		t.Fatalf("world ship leaked SecondaryAmmoPolicy field")
	}
	if len(world.Bullets) != 1 || len(world.Asteroids) != 1 || len(world.Pickups) != 1 {
		t.Fatalf("world projection missing records: %#v", world)
	}

	overlay := ProjectOverlayLane(snapshot, "player-1")
	if overlay.Receiver.SelfID != "player-1" || overlay.Receiver.Lives != 4 {
		t.Fatalf("overlay ownership mismatch: %#v", overlay.Receiver)
	}
	if overlay.Receiver.PrimaryCooldownRemaining != 1.25 || overlay.Receiver.PrimaryAmmoRemaining != 9 {
		t.Fatalf("overlay receiver facts missing: %#v", overlay.Receiver)
	}

	session := ProjectSessionLane(snapshot)
	if len(session.Players) != 1 || session.Players[0].ID != "player-1" {
		t.Fatalf("session players = %#v, want player_sessions ownership", session.Players)
	}
	if session.Players[0].RespawnCooldown != 1.5 || session.TotalAsteroids != 11 {
		t.Fatalf("session projection missing shared facts: %#v", session)
	}
	if len(session.PlayerLifecycle) != 1 || session.PlayerLifecycle[0].Status != "active" {
		t.Fatalf("session lifecycle mismatch: %#v", session.PlayerLifecycle)
	}

	events := ProjectEventLane(snapshot.PendingEvents, 12)
	if len(events.Batch.Events) != 1 || events.Batch.Events[0].EventID != "event-1" {
		t.Fatalf("event batch mismatch: %#v", events.Batch)
	}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1"))
	for _, candidate := range plan.Candidates {
		if candidate.Lane == LaneControl {
			t.Fatalf("planner used slow lane: %#v", candidate)
		}
	}
}

func TestRealtimePlannerUsesGameplayPresentationSnapshotInput(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing"},
		},
		PendingEvents: []game.PendingPresentationEvent{
			{EventID: "event-1", Event: game.EventState{Type: "ship_death"}},
		},
	}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1"))
	if len(plan.Candidates) == 0 {
		t.Fatalf("planner returned no realtime candidates from GameplayPresentationSnapshot input")
	}

	for _, candidate := range plan.Candidates {
		if candidate.Lane == LaneControl {
			t.Fatalf("planner should not depend on Game.StatePacket/statePacket control flow: %#v", candidate)
		}
	}
}
