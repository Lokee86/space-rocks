package tests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestCapsulePolygonCollision(t *testing.T) {
	capsule := game.CollisionBody{
		ID:       "bullet",
		Position: game.Vector2{X: 0, Y: 0},
		Shape:    game.NewCapsuleShape(3, 24),
	}
	polygon := game.CollisionBody{
		ID:       "asteroid",
		Position: game.Vector2{X: 0, Y: 11},
		Shape: game.NewPolygonShape([]game.Vector2{
			{X: -10, Y: -10},
			{X: 10, Y: -10},
			{X: 10, Y: 10},
			{X: -10, Y: 10},
		}),
	}

	if _, ok := game.DetectCollision(capsule, polygon); !ok {
		t.Fatal("expected capsule to collide with polygon")
	}
}

func TestCapsuleConcavePolygonMiss(t *testing.T) {
	capsule := game.CollisionBody{
		ID:       "bullet",
		Position: game.Vector2{X: 0, Y: 0},
		Shape:    game.NewCapsuleShape(2, 8),
	}
	polygon := game.CollisionBody{
		ID:       "asteroid",
		Position: game.Vector2{},
		Shape: game.NewPolygonShape([]game.Vector2{
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

	if _, ok := game.DetectCollision(capsule, polygon); ok {
		t.Fatal("expected capsule to miss concave empty space")
	}
}
