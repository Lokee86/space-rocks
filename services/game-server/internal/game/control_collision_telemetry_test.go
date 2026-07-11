package game

import (
	"math"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestControlCollisionBodiesByKindGroupsAuthoritativeCollisionBodies(t *testing.T) {
	gameInstance := New()
	gameInstance.collisionShapes = physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{Name: "bullet", Type: "circle", Radius: 3},
		Pickups: map[string]physics.ImportedCollisionShape{"powerup": {Name: "powerup", Type: "circle", Radius: 3}},
		Ship: physics.ImportedCollisionShape{Name: "ship", Type: "rectangle", Size: []float64{4, 2}},
	}

	gameInstance.entities.Players["player-1"] = &runtime.Ship{ID: "player-1", X: 10, Y: 20, Rotation: math.Pi / 2}
	gameInstance.entities.Projectiles["bullet-1"] = &runtime.Bullet{ID: "bullet-1", X: 1, Y: 2, Rotation: 0}
	gameInstance.entities.Asteroids["asteroid-1"] = &runtime.Asteroid{ID: "asteroid-1", X: 30, Y: 40, Size: 2, Variant: 0}
	gameInstance.entities.Pickups["pickup-1"] = &pickups.Pickup{ID: "pickup-1", Type: pickups.TypeOneUp, X: -2, Y: 5}
	gameInstance.entities.Pickups["pickup-skipped"] = &pickups.Pickup{ID: "pickup-skipped", Type: pickups.TypeTorpedo, X: 0, Y: 0}

	control := NewControl(gameInstance)
	bodies := control.CollisionBodiesByKind()

	if got := len(bodies["player"]); got != 1 {
		t.Fatalf("expected 1 player body, got %d", got)
	}
	if got := len(bodies["bullet"]); got != 1 {
		t.Fatalf("expected 1 bullet body, got %d", got)
	}
	if got := len(bodies["pickup"]); got != 1 {
		t.Fatalf("expected 1 pickup body, got %d", got)
	}
	if got := len(bodies["asteroid"]); got != 0 {
		t.Fatalf("expected asteroid collision body to be skipped, got %d", got)
	}

	if got := bodies["player"][0].ID; got != "player-1" {
		t.Fatalf("expected player id %q, got %q", "player-1", got)
	}
	if got := bodies["player"][0].Shape.Type; got != "rectangle" {
		t.Fatalf("expected player shape %q, got %q", "rectangle", got)
	}
	if len(bodies["player"][0].Shape.Points) != 4 {
		t.Fatalf("expected raw player outline points, got %d", len(bodies["player"][0].Shape.Points))
	}
}
