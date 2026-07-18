package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
)

func TestFinalMatchStateLocksOnceAndRejectsGameplayMutation(t *testing.T) {
	gameInstance := NewWithSeed(44)
	gameInstance.SetMatchContext("match-1", "trace-1")
	gameInstance.SetModeContext("baseline")
	playerID := gameInstance.AddPlayer()
	if playerID == "" {
		t.Fatal("add player")
	}
	gameInstance.SetPlayerScore(playerID, 42)
	gameInstance.RemovePlayer(playerID)

	locked, ok := gameInstance.LockFinalMatchState()
	if !ok {
		t.Fatal("expected final match state to lock")
	}
	if locked.MatchID != "match-1" || locked.TraceID != "trace-1" || locked.ModeID != "baseline" {
		t.Fatalf("unexpected match identity: %+v", locked)
	}
	if len(locked.Players) != 1 || locked.Players[0].Score != 42 {
		t.Fatalf("unexpected locked player facts: %+v", locked.Players)
	}

	clockAtLock := gameInstance.awardClock
	gameInstance.Step(1)
	if gameInstance.awardClock != clockAtLock {
		t.Fatalf("award clock advanced after lock: before=%v after=%v", clockAtLock, gameInstance.awardClock)
	}
	if added := gameInstance.AddPlayer(); added != "" {
		t.Fatalf("late join succeeded after lock: %q", added)
	}
	if _, err := gameInstance.ApplyGameplayAwardEvent("late-award", []awards.Mutation{{
		Owner: awards.Owner{Scope: awards.ScopePlayer, ID: playerID}, CounterID: awards.CounterScore,
		Operation: awards.MutationIncrement, Value: 100,
	}}); err == nil {
		t.Fatal("award mutation succeeded after lock")
	}

	gameInstance.SetModeContext("changed-after-lock")
	again, ok := gameInstance.LockFinalMatchState()
	if !ok || again.ModeID != "baseline" || again.Players[0].Score != 42 {
		t.Fatalf("locked final state changed: %+v", again)
	}
	again.Players[0].Score = 999
	third, _ := gameInstance.LockedFinalMatchState()
	if third.Players[0].Score != 42 {
		t.Fatalf("returned final state aliased lock: %+v", third.Players)
	}
}
