package physics

import "testing"

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
