package streamruntime

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestContinuousBulletStreamsBeginTracksActiveStream(t *testing.T) {
	streams := NewContinuousBulletStreams()
	origin := physics.Vector2{X: 10, Y: 20}
	direction := physics.Vector2{X: 2, Y: 0}

	if !streams.Begin("player-1", origin, direction) {
		t.Fatal("expected continuous bullet stream to start")
	}

	active := streams.Active()
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

func TestContinuousBulletStreamsRejectInvalidInputs(t *testing.T) {
	streams := NewContinuousBulletStreams()

	if streams.Begin("", physics.Vector2{}, physics.Vector2{X: 1, Y: 0}) {
		t.Fatal("expected empty owner to be rejected")
	}
	if streams.Begin("player-1", physics.Vector2{}, physics.Vector2{}) {
		t.Fatal("expected zero direction to be rejected")
	}
	if got := len(streams.Active()); got != 0 {
		t.Fatalf("expected no active streams, got %d", got)
	}
}

func TestContinuousBulletStreamsClearRemovesStreams(t *testing.T) {
	streams := NewContinuousBulletStreams()
	if !streams.Begin("player-1", physics.Vector2{X: 10, Y: 20}, physics.Vector2{X: 0, Y: -1}) {
		t.Fatal("expected continuous bullet stream to start")
	}

	streams.Clear()

	if got := len(streams.Active()); got != 0 {
		t.Fatalf("expected no active streams, got %d", got)
	}
}

func TestContinuousBulletStreamsStepSpawnsBulletAfterCooldown(t *testing.T) {
	streams := NewContinuousBulletStreams()
	if !streams.Begin("player-1", physics.Vector2{X: 10, Y: 20}, physics.Vector2{X: 0, Y: -1}) {
		t.Fatal("expected continuous bullet stream to start")
	}

	spawnCount := 0
	streams.Step(constants.BasicCannonCooldown, true, func(owner string, origin physics.Vector2, direction physics.Vector2) bool {
		spawnCount++
		if owner != "player-1" {
			t.Fatalf("expected owner %q, got %q", "player-1", owner)
		}
		if origin != (physics.Vector2{X: 10, Y: 20}) {
			t.Fatalf("expected origin %+v, got %+v", physics.Vector2{X: 10, Y: 20}, origin)
		}
		if direction != (physics.Vector2{X: 0, Y: -1}) {
			t.Fatalf("expected direction %+v, got %+v", physics.Vector2{X: 0, Y: -1}, direction)
		}
		return true
	})

	if spawnCount != 1 {
		t.Fatalf("expected 1 spawn, got %d", spawnCount)
	}
	if got := streams.Active()[0].CooldownRemaining; got != constants.BasicCannonCooldown {
		t.Fatalf("expected cooldown reset to %f, got %f", constants.BasicCannonCooldown, got)
	}
}
