package streamruntime

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestRuntimeNewRuntimeBeginsContinuousBulletStream(t *testing.T) {
	runtime := NewRuntime()
	origin := physics.Vector2{X: 10, Y: 20}
	direction := physics.Vector2{X: 2, Y: 0}

	if !runtime.BeginContinuousBulletStream("player-1", origin, direction) {
		t.Fatal("expected continuous bullet stream to start")
	}

	active := runtime.ActiveContinuousBulletStreams()
	if len(active) != 1 {
		t.Fatalf("expected 1 active continuous bullet stream, got %d", len(active))
	}
	if active[0].OwnerPlayerID != "player-1" {
		t.Fatalf("expected owner %q, got %q", "player-1", active[0].OwnerPlayerID)
	}
	if active[0].Origin != origin {
		t.Fatalf("expected origin %+v, got %+v", origin, active[0].Origin)
	}
	if active[0].Direction != (physics.Vector2{X: 1, Y: 0}) {
		t.Fatalf("expected normalized direction, got %+v", active[0].Direction)
	}
	if active[0].CooldownRemaining != constants.BasicCannonCooldown {
		t.Fatalf("expected cooldown %f, got %f", constants.BasicCannonCooldown, active[0].CooldownRemaining)
	}
}

func TestRuntimeClearContinuousBulletStreamsRemovesActiveStreams(t *testing.T) {
	runtime := NewRuntime()
	if !runtime.BeginContinuousBulletStream("player-1", physics.Vector2{X: 10, Y: 20}, physics.Vector2{X: 0, Y: -1}) {
		t.Fatal("expected continuous bullet stream to start")
	}

	runtime.ClearContinuousBulletStreams()

	if got := len(runtime.ActiveContinuousBulletStreams()); got != 0 {
		t.Fatalf("expected no active streams, got %d", got)
	}
}

func TestRuntimeStepContinuousBulletStreamsSpawnsBulletAfterCooldown(t *testing.T) {
	runtime := NewRuntime()
	if !runtime.BeginContinuousBulletStream("player-1", physics.Vector2{X: 10, Y: 20}, physics.Vector2{X: 0, Y: -1}) {
		t.Fatal("expected continuous bullet stream to start")
	}

	spawnCount := 0
	var gotOwner string
	var gotOrigin physics.Vector2
	var gotDirection physics.Vector2

	runtime.StepContinuousBulletStreams(constants.BasicCannonCooldown, true, func(ownerPlayerID string, origin physics.Vector2, direction physics.Vector2) bool {
		spawnCount++
		gotOwner = ownerPlayerID
		gotOrigin = origin
		gotDirection = direction
		return true
	})

	if spawnCount != 1 {
		t.Fatalf("expected 1 spawn, got %d", spawnCount)
	}
	if gotOwner != "player-1" {
		t.Fatalf("expected owner %q, got %q", "player-1", gotOwner)
	}
	if gotOrigin != (physics.Vector2{X: 10, Y: 20}) {
		t.Fatalf("expected origin %+v, got %+v", physics.Vector2{X: 10, Y: 20}, gotOrigin)
	}
	if gotDirection != (physics.Vector2{X: 0, Y: -1}) {
		t.Fatalf("expected direction %+v, got %+v", physics.Vector2{X: 0, Y: -1}, gotDirection)
	}
}
