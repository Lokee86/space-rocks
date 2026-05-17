package game

import "testing"

func TestCapsulePolygonCollision(t *testing.T) {
	capsule := CollisionBody{
		ID:       "bullet",
		Position: Vector2{X: 0, Y: 0},
		Shape:    NewCapsuleShape(3, 24),
	}
	polygon := CollisionBody{
		ID:       "asteroid",
		Position: Vector2{X: 0, Y: 11},
		Shape: NewPolygonShape([]Vector2{
			{X: -10, Y: -10},
			{X: 10, Y: -10},
			{X: 10, Y: 10},
			{X: -10, Y: 10},
		}),
	}

	if _, ok := DetectCollision(capsule, polygon); !ok {
		t.Fatal("expected capsule to collide with polygon")
	}
}

func TestCapsuleConcavePolygonMiss(t *testing.T) {
	capsule := CollisionBody{
		ID:       "bullet",
		Position: Vector2{X: 0, Y: 0},
		Shape:    NewCapsuleShape(2, 8),
	}
	polygon := CollisionBody{
		ID:       "asteroid",
		Position: Vector2{},
		Shape: NewPolygonShape([]Vector2{
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

	if _, ok := DetectCollision(capsule, polygon); ok {
		t.Fatal("expected capsule to miss concave empty space")
	}
}
