package realtime

import (
	"testing"

	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

func TestAssembleRealtimeLaneCandidatesSkipsEventBatchWhenNoPendingEvents(t *testing.T) {
	snapshot := game.GameplayPresentationSnapshot{SelfID: "player-1"}

	plan := AssembleRealtimeLaneCandidates(snapshot, NewRealtimeSessionState("player-1", "match-1"))
	for _, candidate := range plan.Candidates {
		if candidate.Lane() == LaneEvent {
			t.Fatalf("unexpected event lane candidate with no pending events: %#v", candidate)
		}
	}
}
