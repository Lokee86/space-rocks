package tests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestCapsulePolygonCollision(t *testing.T) {
	capsule := physics.CollisionBody{
		ID:       "bullet",
		Position: physics.Vector2{X: 0, Y: 0},
		Shape:    physics.NewCapsuleShape(3, 24),
	}
	polygon := physics.CollisionBody{
		ID:       "asteroid",
		Position: physics.Vector2{X: 0, Y: 11},
		Shape: physics.NewPolygonShape([]physics.Vector2{
			{X: -10, Y: -10},
			{X: 10, Y: -10},
			{X: 10, Y: 10},
			{X: -10, Y: 10},
		}),
	}

	if _, ok := physics.DetectCollision(capsule, polygon); !ok {
		t.Fatal("expected capsule to collide with polygon")
	}
}

func TestCapsuleConcavePolygonMiss(t *testing.T) {
	capsule := physics.CollisionBody{
		ID:       "bullet",
		Position: physics.Vector2{X: 0, Y: 0},
		Shape:    physics.NewCapsuleShape(2, 8),
	}
	polygon := physics.CollisionBody{
		ID:       "asteroid",
		Position: physics.Vector2{},
		Shape: physics.NewPolygonShape([]physics.Vector2{
			{X: -20, Y: -20},
			{X: 20, Y: -20},
			{X: 20, Y: -10},
			{X: -10, Y: -10},
			{X: -10, Y: 10},
			{X: 20, Y: 10},
			{X: 20, Y: 20},
			{X: -20, Y: 20},
		}),
	}

	if _, ok := physics.DetectCollision(capsule, polygon); ok {
		t.Fatal("expected capsule to miss concave empty space")
	}
}
