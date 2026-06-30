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
	state.MarkBaselineReady(LaneWorld)
	state.UpdateLane(LaneSession, Metadata{Sequence: 3, BaselineID: "session-baseline", IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.UpdateLane(LaneEvent, Metadata{Sequence: 9, IsFinalChunk: true})

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if got, want := len(plan.Candidates), 4; got != want {
		t.Fatalf("candidate count = %d, want %d", got, want)
	}

	if got, want := plan.Candidates[0].Lane, Lane(LaneWorld); got != want {
		t.Fatalf("world lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[0].Full.(WorldFullPacket); !ok {
		t.Fatalf("world candidate full type = %T, want WorldFullPacket", plan.Candidates[0].Full)
	} else if got.Metadata.Lane != LaneWorld || len(got.Ships) != 1 || got.Ships[0].ID != "player-1" {
		t.Fatalf("world full packet = %#v, want player-1 ship", got)
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
	if got, ok := plan.Candidates[2].Full.(SessionFullPacket); !ok {
		t.Fatalf("session candidate full type = %T, want SessionFullPacket", plan.Candidates[2].Full)
	} else if got.Metadata.Lane != LaneSession || len(got.Players) != 1 || got.Players[0].ID != "player-1" {
		t.Fatalf("session full packet = %#v, want player-1 session", got)
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
	for _, candidate := range plan.Candidates {
		if packetFamilyForCandidate(candidate) == "" {
			t.Fatalf("expected non-empty packet family for lane=%q kind=%q", candidate.Lane, candidate.Kind)
		}
		wire := WireLanePacket(candidate)
		if gotType, ok := wire["type"].(string); !ok || gotType == "" {
			t.Fatalf("expected top-level type in wired packet for lane=%q kind=%q, got %#v", candidate.Lane, candidate.Kind, wire)
		}
	}
}

func TestAssembleRealtimeLaneCandidatesUsesNextWorldSequenceForUnsyncedFull(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1"}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1"))
	candidate, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatalf("expected world candidate")
	}
	full, ok := candidate.Full.(WorldFullPacket)
	if !ok {
		t.Fatalf("world candidate full type = %T, want WorldFullPacket", candidate.Full)
	}
	if got, want := full.Metadata.Sequence, 1; got != want {
		t.Fatalf("world full sequence = %d, want %d", got, want)
	}
}

func TestAssembleRealtimeLaneCandidatesUsesNextOverlaySequenceForUnsyncedFull(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1", PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1"}}}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1"))
	candidate, ok := findCandidateByLane(plan.Candidates, LaneOverlay)
	if !ok {
		t.Fatalf("expected overlay candidate")
	}
	full, ok := candidate.Full.(OverlayFullPacket)
	if !ok {
		t.Fatalf("overlay candidate full type = %T, want OverlayFullPacket", candidate.Full)
	}
	if got, want := full.Metadata.Sequence, 1; got != want {
		t.Fatalf("overlay full sequence = %d, want %d", got, want)
	}
}

func TestAssembleRealtimeLaneCandidatesUsesNextSessionSequenceForUnsyncedFull(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1", PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1", ShipType: "v_wing"}}}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1"))
	candidate, ok := findCandidateByLane(plan.Candidates, LaneSession)
	if !ok {
		t.Fatalf("expected session candidate")
	}
	full, ok := candidate.Full.(SessionFullPacket)
	if !ok {
		t.Fatalf("session candidate full type = %T, want SessionFullPacket", candidate.Full)
	}
	if got, want := full.Metadata.Sequence, 1; got != want {
		t.Fatalf("session full sequence = %d, want %d", got, want)
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsNextSequenceWorldDeltaForChangedBaseline(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing", X: 12, Y: 34},
		},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, WorldFullPacket{
		Type: PacketTypeWorldFull,
		Metadata: Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true},
		Ships: []WorldShipRecord{{ID: "player-1", ShipType: "v_wing"}},
	})

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	candidate, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatalf("expected world candidate")
	}
	if candidate.Kind != RealtimeLaneCandidateKindDelta {
		t.Fatalf("world candidate kind = %q, want delta", candidate.Kind)
	}
	delta, ok := candidate.Delta.(WorldDeltaPacket)
	if !ok {
		t.Fatalf("world candidate delta type = %T, want WorldDeltaPacket", candidate.Delta)
	}
	if got, want := delta.Metadata.Sequence, 2; got != want {
		t.Fatalf("world delta sequence = %d, want %d", got, want)
	}
	if got, want := delta.Metadata.SnapshotKind, SnapshotKind("delta"); got != want {
		t.Fatalf("world delta snapshot kind = %q, want %q", got, want)
	}
}

func TestAssembleRealtimeLaneCandidatesUsesNextWorldSequenceWhenStoredBaselineExists(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing"}},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	candidate, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatalf("expected world candidate")
	}
	full, ok := candidate.Full.(WorldFullPacket)
	if !ok {
		t.Fatalf("world candidate full type = %T, want WorldFullPacket", candidate.Full)
	}
	if got, want := full.Metadata.Sequence, 2; got != want {
		t.Fatalf("world full sequence = %d, want %d", got, want)
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsValidPacketEnvelopesAfterFinalFullMetadataPersists(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing"},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {
				ID:                "player-1",
				ShipType:          "v_wing",
				Score:             10,
				Lives:             3,
				PrimaryWeaponID:   "laser",
				PrimaryAmmoPolicy: "infinite",
			},
		},
		PlayerLifecycle: map[string]string{"player-1": "active"},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Lane: LaneWorld, Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.UpdateLane(LaneOverlay, Metadata{Lane: LaneOverlay, Sequence: 1, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.UpdateLane(LaneSession, Metadata{Lane: LaneSession, Sequence: 1, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if got, want := len(plan.Candidates), 3; got != want {
		t.Fatalf("candidate count = %d, want %d", got, want)
	}
	for _, candidate := range plan.Candidates {
		if candidate.Kind != RealtimeLaneCandidateKindFull {
			t.Fatalf("expected valid full packet candidate after persisted final full metadata, got lane=%q kind=%q", candidate.Lane, candidate.Kind)
		}
		if packetFamilyForCandidate(candidate) == "" {
			t.Fatalf("expected non-empty packet family for lane=%q kind=%q", candidate.Lane, candidate.Kind)
		}
		wire := WireLanePacket(candidate)
		if gotType, ok := wire["type"].(string); !ok || gotType == "" {
			t.Fatalf("expected top-level type in wired packet for lane=%q kind=%q, got %#v", candidate.Lane, candidate.Kind, wire)
		}
	}
}


func TestAssembleRealtimeLaneCandidatesSkipsEventBatchWhenNoPendingEvents(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
	}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1"))
	for _, candidate := range plan.Candidates {
		if candidate.Lane == LaneEvent {
			t.Fatalf("unexpected event lane candidate with no pending events: %#v", candidate)
		}
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
			t.Fatalf("planner used session lane: %#v", candidate)
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
			t.Fatalf("planner should not depend on old combined packet control flow: %#v", candidate)
		}
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsWorldFullWhenNoBaseline(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing"},
		},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 1, IsFinalChunk: true})

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	world, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatal("expected world candidate when no usable baseline exists")
	}
	if world.Kind != RealtimeLaneCandidateKindFull {
		t.Fatalf("expected world full candidate, got kind=%q", world.Kind)
	}
	if _, ok := world.Full.(WorldFullPacket); !ok {
		t.Fatalf("expected world full packet, got %T", world.Full)
	}
}

func TestAssembleRealtimeLaneCandidatesOmitsWorldWhenStoredBaselineMatches(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing"},
		},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 1, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, BuildWorldFullPacket(snapshot, 1))

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneWorld); ok {
		t.Fatalf("expected no world candidate when stored baseline matches, got %#v", plan.Candidates)
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsWorldDeltaWhenStoredBaselineDiffers(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing", X: 2},
		},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 2, BaselineID: "world-baseline", SnapshotID: "world-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.StoreBaselineProjection(LaneWorld, BuildWorldFullPacket(game.GameplayPresentationSnapshot{SelfID: "player-1", Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing", X: 1}}}, 1))

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	world, ok := findCandidateByLane(plan.Candidates, LaneWorld)
	if !ok {
		t.Fatal("expected world delta candidate when stored baseline differs")
	}
	if world.Kind != RealtimeLaneCandidateKindDelta {
		t.Fatalf("expected world delta candidate, got kind=%q", world.Kind)
	}
	if _, ok := world.Delta.(WorldDeltaPacket); !ok {
		t.Fatalf("expected world delta packet, got %T", world.Delta)
	}
	if _, ok := world.Projection.(WorldFullPacket); !ok {
		t.Fatalf("expected current world full projection to be carried, got %T", world.Projection)
	}
}

func findCandidateByLane(candidates []RealtimeLaneCandidate, lane Lane) (RealtimeLaneCandidate, bool) {
	for _, candidate := range candidates {
		if candidate.Lane == lane {
			return candidate, true
		}
	}
	return RealtimeLaneCandidate{}, false
}

func TestAssembleRealtimeLaneCandidatesOmitsOverlayWhenStoredBaselineMatches(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing"},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 1, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"},
		},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneOverlay, Metadata{Sequence: 4, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.StoreBaselineProjection(LaneOverlay, BuildOverlayFullPacket(snapshot, "player-1", 4))

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneOverlay); ok {
		t.Fatalf("expected no overlay candidate when stored baseline matches, got %#v", plan.Candidates)
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsOverlayDeltaWhenStoredBaselineDiffers(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing"},
		},
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 2, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"},
		},
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneOverlay, Metadata{Sequence: 5, BaselineID: "overlay-baseline", SnapshotID: "overlay-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneOverlay)
	state.StoreBaselineProjection(LaneOverlay, BuildOverlayFullPacket(game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing"}},
		PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1", ShipType: "v_wing", Score: 1, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"}},
	}, "player-1", 4))

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	overlay, ok := findCandidateByLane(plan.Candidates, LaneOverlay)
	if !ok {
		t.Fatal("expected overlay delta candidate when stored baseline differs")
	}
	if overlay.Kind != RealtimeLaneCandidateKindDelta {
		t.Fatalf("expected overlay delta candidate, got kind=%q", overlay.Kind)
	}
	if _, ok := overlay.Delta.(OverlayLaneDelta); !ok {
		t.Fatalf("expected overlay delta packet, got %T", overlay.Delta)
	}
	if _, ok := overlay.Projection.(OverlayFullPacket); !ok {
		t.Fatalf("expected current overlay full projection to be carried, got %T", overlay.Projection)
	}
}

func TestAssembleRealtimeLaneCandidatesOmitsSessionWhenStoredBaselineMatches(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 5, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"},
		},
		PlayerLifecycle: map[string]string{"player-1": "active"},
		TotalAsteroids: 5,
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneSession, Metadata{Sequence: 7, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.StoreBaselineProjection(LaneSession, BuildSessionFullPacket(snapshot, 7))

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if _, ok := findCandidateByLane(plan.Candidates, LaneSession); ok {
		t.Fatalf("expected no session candidate when stored baseline matches, got %#v", plan.Candidates)
	}
}

func TestAssembleRealtimeLaneCandidatesEmitsSessionDeltaWhenStoredBaselineDiffers(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		PlayerSessions: map[string]game.PlayerSessionState{
			"player-1": {ID: "player-1", ShipType: "v_wing", Score: 9, Lives: 2, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"},
		},
		PlayerLifecycle: map[string]string{"player-1": "respawning"},
		TotalAsteroids: 8,
	}

	state := NewRealtimeSessionState("player-1")
	state.UpdateLane(LaneSession, Metadata{Sequence: 8, BaselineID: "session-baseline", SnapshotID: "session-baseline", SnapshotKind: SnapshotKind("full"), IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.StoreBaselineProjection(LaneSession, BuildSessionFullPacket(game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1", ShipType: "v_wing", Score: 5, Lives: 3, PrimaryWeaponID: "laser", PrimaryAmmoPolicy: "infinite"}},
		PlayerLifecycle: map[string]string{"player-1": "active"},
		TotalAsteroids: 5,
	}, 7))

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	session, ok := findCandidateByLane(plan.Candidates, LaneSession)
	if !ok {
		t.Fatal("expected session delta candidate when stored baseline differs")
	}
	if session.Kind != RealtimeLaneCandidateKindDelta {
		t.Fatalf("expected session delta candidate, got kind=%q", session.Kind)
	}
	if _, ok := session.Delta.(SessionLaneDelta); !ok {
		t.Fatalf("expected session delta packet, got %T", session.Delta)
	}
	if _, ok := session.Projection.(SessionFullPacket); !ok {
		t.Fatalf("expected current session full projection to be carried, got %T", session.Projection)
	}
}
