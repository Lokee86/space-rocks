package bots

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func TestControllerTurnsAndFiresTowardAsteroid(t *testing.T) {
	controller := NewController()
	input := controller.Decide(Observation{
		Position:  physics.Vector2{X: 100, Y: 100},
		Rotation:  0,
		Asteroids: []AsteroidObservation{{Position: physics.Vector2{X: 100, Y: 20}}},
	})

	if !input.Forward {
		t.Fatal("expected bot to thrust toward asteroid")
	}
	if !input.PrimaryFire {
		t.Fatal("expected bot to fire while asteroid is aligned")
	}
	if input.Left || input.Right {
		t.Fatalf("expected no turn for aligned asteroid, got %+v", input)
	}
}

func TestControllerAvoidsImminentCollision(t *testing.T) {
	controller := NewController()
	input := controller.Decide(Observation{
		Position: physics.Vector2{X: 100, Y: 100},
		Velocity: physics.Vector2{X: 0, Y: -100},
		Rotation: 0,
		Asteroids: []AsteroidObservation{{
			Position: physics.Vector2{X: 120, Y: 20},
			Velocity: physics.Vector2{X: 0, Y: 0},
			Size:     3,
		}},
	})

	if !input.Left && !input.Right {
		t.Fatalf("expected avoidance turn, got %+v", input)
	}
}
