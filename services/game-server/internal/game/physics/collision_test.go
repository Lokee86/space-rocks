package physics

import "testing"

func TestBodyContainsPointCircle(t *testing.T) {
	body := CollisionBody{
		Position: Vector2{X: 10, Y: 20},
		Shape:    NewCircleShape(5),
	}

	if !BodyContainsPoint(body, Vector2{X: 13, Y: 24}) {
		t.Fatalf("expected point to be inside circle body")
	}
	if BodyContainsPoint(body, Vector2{X: 16, Y: 20}) {
		t.Fatalf("expected point to be outside circle body")
	}
}

func TestBodyContainsPointCapsule(t *testing.T) {
	body := CollisionBody{
		Position: Vector2{X: 0, Y: 0},
		Rotation: 0,
		Shape:    NewCapsuleShape(2, 10),
	}

	if !BodyContainsPoint(body, Vector2{X: 1, Y: 0}) {
		t.Fatalf("expected point to be inside capsule body")
	}
	if BodyContainsPoint(body, Vector2{X: 3, Y: 0}) {
		t.Fatalf("expected point to be outside capsule body")
	}
}

func TestBodyContainsPointRectangle(t *testing.T) {
	body := CollisionBody{
		Position: Vector2{X: 5, Y: 5},
		Shape:    NewRectangleShape(8, 4),
	}

	if !BodyContainsPoint(body, Vector2{X: 7, Y: 6}) {
		t.Fatalf("expected point to be inside rectangle body")
	}
	if BodyContainsPoint(body, Vector2{X: 10, Y: 8}) {
		t.Fatalf("expected point to be outside rectangle body")
	}
}

func TestBodyContainsPointPolygon(t *testing.T) {
	body := CollisionBody{
		Position: Vector2{X: 0, Y: 0},
		Shape: NewPolygonShape([]Vector2{
			{X: 0, Y: 0},
			{X: 4, Y: 0},
			{X: 2, Y: 4},
		}),
	}

	if !BodyContainsPoint(body, Vector2{X: 2, Y: 1}) {
		t.Fatalf("expected point to be inside polygon body")
	}
	if BodyContainsPoint(body, Vector2{X: 3.5, Y: 3.5}) {
		t.Fatalf("expected point to be outside polygon body")
	}
}
