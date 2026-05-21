package physicstests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestLoadCollisionShapeCatalog(t *testing.T) {
	catalog, err := physics.LoadCollisionShapeCatalog()
	if err != nil {
		t.Fatal(err)
	}

	if catalog.Bullet.Type != "capsule" {
		t.Fatalf("expected bullet capsule shape, got %q", catalog.Bullet.Type)
	}
	if catalog.Ship.Type != "polygon" {
		t.Fatalf("expected ship polygon shape, got %q", catalog.Ship.Type)
	}

	if len(catalog.Asteroids) != 4 {
		t.Fatalf("expected 4 asteroid collision variants, got %d", len(catalog.Asteroids))
	}
}

func TestAsteroidShapeScalesImportedPolygon(t *testing.T) {
	catalog := physics.CollisionShapeCatalog{
		Asteroids: []physics.ImportedCollisionShape{
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

	if shape.Type != physics.CollisionShapePolygon {
		t.Fatalf("expected polygon shape, got %s", shape.Type)
	}
	if shape.Points[0].X != 2 {
		t.Fatalf("expected first point X to scale to 2, got %v", shape.Points[0].X)
	}
}

func TestShipShapeByIDReturnsDefaultShipShape(t *testing.T) {
	catalog := testShipShapeCatalog()

	defaultShape, err := catalog.ShipShape()
	if err != nil {
		t.Fatal(err)
	}
	shape, err := catalog.ShipShapeByID(physics.DefaultShipCollisionShapeID)
	if err != nil {
		t.Fatal(err)
	}

	assertSameCircleShape(t, shape, defaultShape)
}

func TestShipShapeByIDFallsBackForUnknownID(t *testing.T) {
	catalog := testShipShapeCatalog()

	defaultShape, err := catalog.ShipShape()
	if err != nil {
		t.Fatal(err)
	}
	shape, err := catalog.ShipShapeByID("unknown_ship")
	if err != nil {
		t.Fatal(err)
	}

	assertSameCircleShape(t, shape, defaultShape)
}

func testShipShapeCatalog() physics.CollisionShapeCatalog {
	return physics.CollisionShapeCatalog{
		Ship: physics.ImportedCollisionShape{
			Type:   "circle",
			Radius: 20,
		},
	}
}

func assertSameCircleShape(t *testing.T, shape physics.CollisionShape, expected physics.CollisionShape) {
	t.Helper()

	if shape.Type != expected.Type || shape.Radius != expected.Radius {
		t.Fatalf("expected circle shape %#v, got %#v", expected, shape)
	}
}
