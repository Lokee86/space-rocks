package spacetests

import (
	"testing"

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

func TestNormalizePositionIsNoOp(t *testing.T) {
	position := physics.Vector2{X: -12, Y: 34}
	normalized := space.NormalizePosition(position)

	if normalized != position {
		t.Fatalf("expected normalize position to return %+v, got %+v", position, normalized)
	}
}
