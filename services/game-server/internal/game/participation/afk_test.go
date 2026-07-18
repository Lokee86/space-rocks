package participation

import (
	"testing"

	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
)

func TestAFKRuntimeKeepsOneTimerAcrossActiveAndPendingRespawn(t *testing.T) {
	runtime, err := NewRuntime(AFKPolicy{Timeout: 35})
	if err != nil {
		t.Fatal(err)
	}
	if err := runtime.RegisterParticipant("player-1"); err != nil {
		t.Fatal(err)
	}
	status := playerstate.StatusActive
	lookup := func(string) (playerstate.Status, bool) { return status, true }
	if got := runtime.Step(20, lookup); len(got) != 0 {
		t.Fatalf("unexpected expiry: %+v", got)
	}
	status = playerstate.StatusPendingRespawn
	if got := runtime.Step(14, lookup); len(got) != 0 {
		t.Fatalf("unexpected expiry before allowance: %+v", got)
	}
	got := runtime.Step(1, lookup)
	if len(got) != 1 || got[0].ReasonCode != "afk_forfeit" {
		t.Fatalf("unexpected expiry: %+v", got)
	}
	if got := runtime.Step(1, lookup); len(got) != 0 {
		t.Fatalf("expiry repeated: %+v", got)
	}
}

func TestAFKRuntimeActionResetsAndInactiveStatusDoesNotTick(t *testing.T) {
	runtime, _ := NewRuntime(NewDefaultAFKPolicy())
	_ = runtime.RegisterParticipant("player-1")
	status := playerstate.StatusEliminated
	lookup := func(string) (playerstate.Status, bool) { return status, true }
	if got := runtime.Step(100, lookup); len(got) != 0 {
		t.Fatalf("inactive participant expired: %+v", got)
	}
	status = playerstate.StatusActive
	if !runtime.RecordAction("player-1") {
		t.Fatal("expected action to reset timer")
	}
	if got := runtime.Step(34, lookup); len(got) != 0 {
		t.Fatalf("action reset did not restore allowance: %+v", got)
	}
}
