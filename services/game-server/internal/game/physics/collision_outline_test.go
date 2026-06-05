package physics

import (
	"math"
	"testing"
)

func TestCollisionBodyOutlinePointsPolygonRotationTranslation(t *testing.T) {
	body := CollisionBody{
		Position: Vector2{X: 10, Y: -4},
		Rotation: math.Pi / 2,
		Shape: NewPolygonShape([]Vector2{
			{X: -2, Y: -1},
			{X: 2, Y: -1},
			{X: 2, Y: 1},
			{X: -2, Y: 1},
		}),
	}

	points := CollisionBodyOutlinePoints(body)
	expected := []Vector2{
		{X: 11, Y: -6},
		{X: 11, Y: -2},
		{X: 9, Y: -2},
		{X: 9, Y: -6},
	}

	if len(points) != len(expected) {
		t.Fatalf("expected %d points, got %d", len(expected), len(points))
	}

	for index, point := range points {
		assertVectorApproxEqual(t, point, expected[index])
	}
}

func TestCollisionBodyOutlinePointsCapsuleOrientation(t *testing.T) {
	body := CollisionBody{
		Position: Vector2{X: 0, Y: 0},
		Rotation: math.Pi / 2,
		Shape:    NewCapsuleShape(2, 10),
	}

	points := CollisionBodyOutlinePoints(body)
	if len(points) != collisionOutlineCircleSegments+2 {
		t.Fatalf("expected %d points, got %d", collisionOutlineCircleSegments+2, len(points))
	}

	minX, maxX := points[0].X, points[0].X
	minY, maxY := points[0].Y, points[0].Y
	for _, point := range points[1:] {
		minX = math.Min(minX, point.X)
		maxX = math.Max(maxX, point.X)
		minY = math.Min(minY, point.Y)
		maxY = math.Max(maxY, point.Y)
	}

	assertFloatApproxEqual(t, minX, -3)
	assertFloatApproxEqual(t, maxX, 3)
	assertFloatApproxEqual(t, minY, -2)
	assertFloatApproxEqual(t, maxY, 2)
}

func assertVectorApproxEqual(t *testing.T, actual Vector2, expected Vector2) {
	t.Helper()

	assertFloatApproxEqual(t, actual.X, expected.X)
	assertFloatApproxEqual(t, actual.Y, expected.Y)
}

func assertFloatApproxEqual(t *testing.T, actual float64, expected float64) {
	t.Helper()

	const epsilon = 1e-9
	if math.Abs(actual-expected) > epsilon {
		t.Fatalf("expected %.9f, got %.9f", expected, actual)
	}
}
