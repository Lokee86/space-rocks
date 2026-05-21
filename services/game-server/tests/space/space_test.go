package spacetests

import (
	"math"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/space"
)

func TestDeltaUsesFlatWorldDifference(t *testing.T) {
	delta := space.Delta(physics.Vector2{}, physics.Vector2{X: 3, Y: 4})

	if delta.X != 3 || delta.Y != 4 {
		t.Fatalf("expected delta (3, 4), got (%v, %v)", delta.X, delta.Y)
	}
}

func TestDistanceUsesFlatWorldLength(t *testing.T) {
	distance := space.Distance(physics.Vector2{}, physics.Vector2{X: 3, Y: 4})

	if distance != 5 {
		t.Fatalf("expected distance 5, got %v", distance)
	}
}

func TestDirectionUsesNormalizedFlatWorldDelta(t *testing.T) {
	direction := space.Direction(physics.Vector2{}, physics.Vector2{X: 3, Y: 4})

	if direction.X != 0.6 || direction.Y != 0.8 {
		t.Fatalf("expected direction (0.6, 0.8), got (%v, %v)", direction.X, direction.Y)
	}
}

func TestDirectionUsesShortestWrappedPath(t *testing.T) {
	direction := space.Direction(
		physics.Vector2{X: constants.WorldWidth - 5, Y: 10},
		physics.Vector2{X: 5, Y: 10},
	)

	if direction != (physics.Vector2{X: 1, Y: 0}) {
		t.Fatalf("expected wrapped direction (1, 0), got %+v", direction)
	}
}

func TestWrapPositionWrapsRightEdgeToLeft(t *testing.T) {
	wrapped := space.WrapPosition(physics.Vector2{X: 105, Y: 25}, space.Bounds{Width: 100, Height: 50})

	if wrapped != (physics.Vector2{X: 5, Y: 25}) {
		t.Fatalf("expected right edge wrap to (5, 25), got %+v", wrapped)
	}
}

func TestWrapPositionWrapsLeftEdgeToRight(t *testing.T) {
	wrapped := space.WrapPosition(physics.Vector2{X: -5, Y: 25}, space.Bounds{Width: 100, Height: 50})

	if wrapped != (physics.Vector2{X: 95, Y: 25}) {
		t.Fatalf("expected left edge wrap to (95, 25), got %+v", wrapped)
	}
}

func TestWrapPositionWrapsBottomEdgeToTop(t *testing.T) {
	wrapped := space.WrapPosition(physics.Vector2{X: 25, Y: 55}, space.Bounds{Width: 100, Height: 50})

	if wrapped != (physics.Vector2{X: 25, Y: 5}) {
		t.Fatalf("expected bottom edge wrap to (25, 5), got %+v", wrapped)
	}
}

func TestWrapPositionWrapsTopEdgeToBottom(t *testing.T) {
	wrapped := space.WrapPosition(physics.Vector2{X: 25, Y: -5}, space.Bounds{Width: 100, Height: 50})

	if wrapped != (physics.Vector2{X: 25, Y: 45}) {
		t.Fatalf("expected top edge wrap to (25, 45), got %+v", wrapped)
	}
}

func TestWrapPositionHandlesPositionsMoreThanOneWorldSizeOut(t *testing.T) {
	wrapped := space.WrapPosition(physics.Vector2{X: 235, Y: -125}, space.Bounds{Width: 100, Height: 50})

	if wrapped != (physics.Vector2{X: 35, Y: 25}) {
		t.Fatalf("expected multi-world wrap to (35, 25), got %+v", wrapped)
	}
}

func TestShortestDeltaCrossesHorizontalEdge(t *testing.T) {
	delta := space.ShortestDelta(
		physics.Vector2{X: 95, Y: 25},
		physics.Vector2{X: 5, Y: 25},
		space.Bounds{Width: 100, Height: 50},
	)

	if delta != (physics.Vector2{X: 10, Y: 0}) {
		t.Fatalf("expected horizontal wrapped delta (10, 0), got %+v", delta)
	}
}

func TestShortestDeltaCrossesVerticalEdge(t *testing.T) {
	delta := space.ShortestDelta(
		physics.Vector2{X: 25, Y: 45},
		physics.Vector2{X: 25, Y: 5},
		space.Bounds{Width: 100, Height: 50},
	)

	if delta != (physics.Vector2{X: 0, Y: 10}) {
		t.Fatalf("expected vertical wrapped delta (0, 10), got %+v", delta)
	}
}

func TestShortestDeltaStaysDirectWhenDirectPathIsShorter(t *testing.T) {
	delta := space.ShortestDelta(
		physics.Vector2{X: 10, Y: 10},
		physics.Vector2{X: 35, Y: 30},
		space.Bounds{Width: 100, Height: 50},
	)

	if delta != (physics.Vector2{X: 25, Y: 20}) {
		t.Fatalf("expected direct delta (25, 20), got %+v", delta)
	}
}

func TestWrappedDistanceUsesShortestWrappedPath(t *testing.T) {
	distance := space.WrappedDistance(
		physics.Vector2{X: 95, Y: 45},
		physics.Vector2{X: 5, Y: 5},
		space.Bounds{Width: 100, Height: 50},
	)
	expected := math.Sqrt(200)

	if distance != expected {
		t.Fatalf("expected wrapped distance %v, got %v", expected, distance)
	}
}

func TestNormalizePositionWrapsIntoDefaultBounds(t *testing.T) {
	position := physics.Vector2{X: -12, Y: 34}
	normalized := space.NormalizePosition(position)
	expected := physics.Vector2{X: constants.WorldWidth - 12, Y: 34}

	if normalized != expected {
		t.Fatalf("expected normalize position to return %+v, got %+v", expected, normalized)
	}
}

func TestNormalizePositionWrapsBothAxesIntoDefaultBounds(t *testing.T) {
	position := physics.Vector2{X: constants.WorldWidth + 5, Y: -3}
	normalized := space.NormalizePosition(position)
	expected := physics.Vector2{X: 5, Y: constants.WorldHeight - 3}

	if normalized != expected {
		t.Fatalf("expected normalize position to return %+v, got %+v", expected, normalized)
	}
}
