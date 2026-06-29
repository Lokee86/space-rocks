package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/server/internal/game"
)

func TestRealtimeEventBatchProjectionSchedulingEncodingDoNotDrain(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{
		SelfID: "player-1",
		PendingEvents: []game.PendingPresentationEvent{
			{EventID: "event-1", Event: game.EventState{Type: "ship_death"}},
		},
	}

	state := NewRealtimeSessionState("player-1")
	plan := AssembleRealtimeLaneCandidates(snapshot, state)
	if len(snapshot.PendingEvents) != 1 {
		t.Fatalf("projection mutated pending events: %d", len(snapshot.PendingEvents))
	}
	if len(plan.Candidates) != 4 {
		t.Fatalf("unexpected candidate count: %d", len(plan.Candidates))
	}
	if len(snapshot.PendingEvents) != 1 {
		t.Fatalf("scheduling mutated pending events: %d", len(snapshot.PendingEvents))
	}

	active := BuildActiveRealtimeResult(snapshot, state)
	if len(snapshot.PendingEvents) != 1 {
		t.Fatalf("encoding mutated pending events: %d", len(snapshot.PendingEvents))
	}
	if len(active.EventBatchEventIDs) != 1 || active.EventBatchEventIDs[0] != "event-1" {
		t.Fatalf("expected active result to retain event IDs, got %#v", active.EventBatchEventIDs)
	}
}
