package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

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

	state := NewRealtimeSessionState("player-1", "match-1")
	state.UpdateLane(LaneWorld, Metadata{Sequence: 1, BaselineID: "world-baseline", IsFinalChunk: true})
	state.MarkBaselineReady(LaneWorld)
	state.UpdateLane(LaneSession, Metadata{Sequence: 3, BaselineID: "session-baseline", IsFinalChunk: true})
	state.MarkBaselineReady(LaneSession)
	state.UpdateLane(LaneEvent, Metadata{Sequence: 9, IsFinalChunk: true})

	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if got, want := len(plan.Candidates), 4; got != want {
		t.Fatalf("candidate count = %d, want %d", got, want)
	}

	if got, want := plan.Candidates[0].Lane(), Lane(LaneWorld); got != want {
		t.Fatalf("world lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[0].Payload.(WorldWireFullPacket); !ok {
		t.Fatalf("world candidate full type = %T, want WorldWireFullPacket", plan.Candidates[0].Payload)
	} else if got.Metadata.Lane != LaneWorld || len(got.Ships) != 1 || got.Ships[0].ID != "player-1" {
		t.Fatalf("world full packet = %#v, want player-1 ship", got)
	}

	if got, want := plan.Candidates[1].Lane(), Lane(LaneOverlay); got != want {
		t.Fatalf("overlay lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[1].Payload.(OverlayWireFullPacket); !ok {
		t.Fatalf("overlay candidate full type = %T, want OverlayWireFullPacket", plan.Candidates[1].Payload)
	} else if got.Metadata.Lane != LaneOverlay || got.Receiver.SelfID != "player-1" {
		t.Fatalf("overlay packet = %#v, want player-1 overlay packet", got)
	}

	if got, want := plan.Candidates[2].Lane(), Lane(LaneSession); got != want {
		t.Fatalf("session lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[2].Payload.(SessionWireFullPacket); !ok {
		t.Fatalf("session candidate full type = %T, want SessionWireFullPacket", plan.Candidates[2].Payload)
	} else if got.Metadata.Lane != LaneSession || len(got.Players) != 1 || got.Players[0].ID != "player-1" {
		t.Fatalf("session full packet = %#v, want player-1 session", got)
	}

	if got, want := plan.Candidates[3].Lane(), Lane(LaneEvent); got != want {
		t.Fatalf("event lane = %q, want %q", got, want)
	}
	if got, ok := plan.Candidates[3].Payload.(EventBatchPacket); !ok {
		t.Fatalf("event candidate full type = %T, want EventBatchPacket", plan.Candidates[3].Payload)
	} else if got.Metadata.Lane != LaneEvent || len(got.Batch.Events) != 1 || got.Batch.Events[0].EventID != "event-1" {
		t.Fatalf("event batch = %#v, want preserved event id", got)
	}

	if len(snapshot.PendingEvents) != 1 || snapshot.PendingEvents[0].EventID != "event-1" {
		t.Fatalf("planner mutated pending events: %#v", snapshot.PendingEvents)
	}
	for _, candidate := range plan.Candidates {
		if candidate.PacketFamily() == "" {
			t.Fatalf("expected non-empty packet family for lane=%q kind=%q", candidate.Lane(), candidate.Kind())
		}
		wire := mustWireLanePacket(t, candidate)
		if gotType, ok := wire["type"].(string); !ok || gotType == "" {
			t.Fatalf("expected top-level type in wired packet for lane=%q kind=%q, got %#v", candidate.Lane(), candidate.Kind(), wire)
		}
	}
}

func TestAssembleRealtimeLaneCandidatesUsesNextOverlaySequenceForUnsyncedFull(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1", PlayerSessions: map[string]game.PlayerSessionState{"player-1": {ID: "player-1"}}}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1", "match-1"))
	candidate, ok := findCandidateByLane(plan.Candidates, LaneOverlay)
	if !ok {
		t.Fatalf("expected overlay candidate")
	}
	full, ok := candidate.Payload.(OverlayWireFullPacket)
	if !ok {
		t.Fatalf("overlay candidate full type = %T, want OverlayWireFullPacket", candidate.Payload)
	}
	if got, want := full.Metadata.Sequence, 1; got != want {
		t.Fatalf("overlay full sequence = %d, want %d", got, want)
	}
}