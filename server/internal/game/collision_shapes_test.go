package game

import "testing"

func TestLoadCollisionShapeCatalog(t *testing.T) {
	catalog, err := LoadCollisionShapeCatalog()
	if err != nil {
		t.Fatal(err)
	}

	if catalog.Bullet.Type != "capsule" {
		t.Fatalf("expected bullet capsule shape, got %q", catalog.Bullet.Type)
	}

	if len(catalog.Asteroids) != 4 {
		t.Fatalf("expected 4 asteroid collision variants, got %d", len(catalog.Asteroids))
	}
}

func TestAsteroidShapeScalesImportedPolygon(t *testing.T) {
	catalog := CollisionShapeCatalog{
		Asteroids: []ImportedCollisionShape{
			{
				Type: "polygon",
				Points: [][]float64{
					{8, 0},
					{0, 8},
					{-8, 0},
				},
			},
		},
	}

	shape, err := catalog.AsteroidShape(0, 2)
	if err != nil {
		t.Fatal(err)
	}

	if shape.Type != CollisionShapePolygon {
		t.Fatalf("expected polygon shape, got %s", shape.Type)
	}
	if shape.Points[0].X != 2 {
		t.Fatalf("expected first point X to scale to 2, got %v", shape.Points[0].X)
	}
}
