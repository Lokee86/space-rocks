package physics

import (
	"math"
	"testing"
)

func TestBoundingRadius(t *testing.T) {
	tests := []struct {
		name  string
		shape CollisionShape
		want  float64
	}{
		{name: "circle", shape: NewCircleShape(3), want: 3},
		{name: "capsule", shape: NewCapsuleShape(2, 10), want: 5},
		{name: "capsule uses radius", shape: NewCapsuleShape(7, 4), want: 7},
		{name: "rectangle", shape: NewRectangleShape(6, 8), want: 5},
		{name: "polygon", shape: NewPolygonShape([]Vector2{{X: 1, Y: 2}, {X: -4, Y: 3}, {X: 2, Y: 1}}), want: 5},
		{name: "empty", shape: CollisionShape{}, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := BoundingRadius(test.shape); math.Abs(got-test.want) > 1e-9 {
				t.Fatalf("BoundingRadius() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCollisionShapeCatalogPickupShapeLoadsOneUp(t *testing.T) {
	catalog := CollisionShapeCatalog{
		Pickups: map[string]ImportedCollisionShape{
			"1_up": {
				Name:   "CollisionShape2D",
				Type:   "circle",
				Radius: 50,
			},
		},
	}

	shape, err := catalog.PickupShape("1_up")
	if err != nil {
		t.Fatalf("PickupShape() error = %v", err)
	}
	if shape.Type != CollisionShapeCircle {
		t.Fatalf("PickupShape().Type = %q, want %q", shape.Type, CollisionShapeCircle)
	}
	if shape.Radius != 50 {
		t.Fatalf("PickupShape().Radius = %v, want 50", shape.Radius)
	}
}

func TestCollisionShapeCatalogPickupShapeMissingTypeReturnsError(t *testing.T) {
	catalog := CollisionShapeCatalog{
		Pickups: map[string]ImportedCollisionShape{
			"1_up": {
				Name:   "CollisionShape2D",
				Type:   "circle",
				Radius: 50,
			},
		},
	}

	_, err := catalog.PickupShape("missing")
	if err == nil {
		t.Fatal("PickupShape() error = nil, want error")
	}
}

func TestCollisionShapeCatalogPickupShapeMissingCatalogReturnsError(t *testing.T) {
	catalog := CollisionShapeCatalog{}

	_, err := catalog.PickupShape("1_up")
	if err == nil {
		t.Fatal("PickupShape() error = nil, want error")
	}
}

func TestCollisionShapeCatalogPickupShapeUsesImportedRadius(t *testing.T) {
	catalog := CollisionShapeCatalog{
		Pickups: map[string]ImportedCollisionShape{
			"1_up": {
				Name:   "CollisionShape2D",
				Type:   "circle",
				Radius: 12.5,
			},
		},
	}

	shape, err := catalog.PickupShape("1_up")
	if err != nil {
		t.Fatalf("PickupShape() error = %v", err)
	}
	if shape.Radius != 12.5 {
		t.Fatalf("PickupShape().Radius = %v, want 12.5", shape.Radius)
	}
}
