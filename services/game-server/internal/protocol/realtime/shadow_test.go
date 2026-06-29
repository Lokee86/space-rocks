package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestBuildShadowRealtimeResultDoesNotDrainPendingEvents(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{
			"player-1": {ID: "player-1", ShipType: "v_wing"},
		},
		PendingEvents: []game.PendingPresentationEvent{{EventID: "event-1", Event: game.EventState{Type: "ship_death"}}},
	}
	state := NewRealtimeSessionState("player-1")

	result := BuildShadowRealtimeResult(snapshot, state)
	if len(result.Candidates) == 0 {
		t.Fatalf("shadow result has no candidates")
	}

	var eventCandidateFound bool
	for _, candidate := range result.Candidates {
		if candidate.Lane != LaneEvent {
			continue
		}
		eventCandidateFound = true
		batch, ok := candidate.Full.(EventBatchPacket)
		if !ok {
			t.Fatalf("event candidate full type = %T, want EventBatchPacket", candidate.Full)
		}
		if len(batch.Batch.Events) != 1 || batch.Batch.Events[0].EventID != "event-1" {
			t.Fatalf("event batch = %#v, want preserved event id", batch)
		}
		if got := result.EncodedBytes[LaneEvent]; got == 0 {
			t.Fatalf("expected encoded bytes to be recorded for event lane")
		}
	}
	if !eventCandidateFound {
		t.Fatalf("shadow result did not include event batch candidate")
	}

	if len(snapshot.PendingEvents) != 1 || snapshot.PendingEvents[0].EventID != "event-1" {
		t.Fatalf("shadow builder mutated pending events: %#v", snapshot.PendingEvents)
	}
	if len(snapshot.PendingEvents) != 1 {
		t.Fatalf("expected pending events to remain queued after shadow build")
	}
}


func TestShadowRealtimeResultKeepsEventIDsStableFromProjectionToOutput(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		Players: map[string]runtime.ShipState{"player-1": {ID: "player-1", ShipType: "v_wing"}},
		PendingEvents: []game.PendingPresentationEvent{{EventID: "event-1", Event: game.EventState{Type: "ship_death"}}},
	}

	projection := ProjectEventLane(snapshot.PendingEvents, 1)
	if len(projection.Batch.Events) != 1 || projection.Batch.Events[0].EventID != "event-1" {
		t.Fatalf("projection event IDs = %#v", projection.Batch.Events)
	}

	result := BuildShadowRealtimeResult(snapshot, NewRealtimeSessionState("player-1"))
	if len(result.Candidates) == 0 {
		t.Fatalf("shadow result has no candidates")
	}
	for _, candidate := range result.Candidates {
		if candidate.Lane != LaneEvent {
			continue
		}
		batch, ok := candidate.Full.(EventBatchPacket)
		if !ok || len(batch.Batch.Events) != 1 || batch.Batch.Events[0].EventID != "event-1" {
			t.Fatalf("shadow output event IDs = %#v", candidate.Full)
		}
	}
}
