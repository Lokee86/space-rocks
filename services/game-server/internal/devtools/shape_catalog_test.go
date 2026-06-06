package devtools

import (
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestBuildShapeCatalogIncludesExpectedShapeIDs(t *testing.T) {
	catalog := physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{
			Name:   "bullet",
			Type:   string(physics.CollisionShapeCircle),
			Radius: 1,
		},
		Ship: physics.ImportedCollisionShape{
			Name:   "ship",
			Type:   string(physics.CollisionShapePolygon),
			Points: [][]float64{{-1, 0}, {0, 1}, {1, 0}},
		},
		Asteroids: []physics.ImportedCollisionShape{
			{
				Name:   "asteroid-0",
				Type:   string(physics.CollisionShapePolygon),
				Points: [][]float64{{-1, -1}, {1, -1}, {0, 1}},
			},
		},
		Pickups: map[string]physics.ImportedCollisionShape{
			"1_up": {
				Name:   "pickup-1-up",
				Type:   string(physics.CollisionShapeCircle),
				Radius: 0.5,
			},
		},
	}

	got := BuildShapeCatalog(catalog)

	assertShapeEntry(t, got, "player:v_wing", "player")
	assertShapeEntry(t, got, "bullet", "bullet")
	assertShapeEntry(t, got, "asteroid:0", "asteroid")
	assertShapeEntry(t, got, "pickup:1_up", "pickup")
}

func TestBuildShapeCatalogSkipsInvalidShapes(t *testing.T) {
	catalog := physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{
			Name: "invalid-bullet",
			Type: "circle",
		},
	}

	got := BuildShapeCatalog(catalog)

	if _, ok := got["bullet"]; ok {
		t.Fatalf("BuildShapeCatalog() included invalid bullet shape")
	}
}

func assertShapeEntry(t *testing.T, got map[string]DebugShapeDefinition, id string, kind string) {
	t.Helper()

	entry, ok := got[id]
	if !ok {
		t.Fatalf("BuildShapeCatalog() missing %q", id)
	}
	if entry.ID != id {
		t.Fatalf("entry.ID = %q, want %q", entry.ID, id)
	}
	if entry.Kind != kind {
		t.Fatalf("entry.Kind = %q, want %q", entry.Kind, kind)
	}
	if len(entry.Points) == 0 {
		t.Fatalf("entry.Points is empty for %q", id)
	}
	if strings.Contains(entry.ID, "player-") || strings.Contains(entry.ID, "asteroid-") || strings.Contains(entry.ID, "pickup-") {
		t.Fatalf("entry.ID %q unexpectedly uses a live entity id format", entry.ID)
	}
	for _, point := range entry.Points {
		if point.X == 0 && point.Y == 0 {
			t.Fatalf("entry.Points for %q unexpectedly contains origin position", id)
		}
	}
}
