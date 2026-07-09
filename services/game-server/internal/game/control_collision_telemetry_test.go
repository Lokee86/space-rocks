package game

import (
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestDevtoolsCollisionBodiesUsesServerCollisionBodies(t *testing.T) {
	gameInstance := New()
	gameInstance.collisionShapes = physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{ Name: "bullet", Type: "circle", Radius: 3 },
		Pickups: map[string]physics.ImportedCollisionShape{"powerup": {Name: "powerup", Type: "circle", Radius: 3}},
		Ship: physics.ImportedCollisionShape{Name: "ship", Type: "rectangle", Size: []float64{4, 2}},
	}

	gameInstance.entities.Players["player-1"] = &runtime.Ship{ID: "player-1", X: 10, Y: 20, Rotation: math.Pi / 2}
	gameInstance.entities.Projectiles["bullet-1"] = &runtime.Bullet{ID: "bullet-1", X: 1, Y: 2, Rotation: 0}
	gameInstance.entities.Asteroids["asteroid-1"] = &runtime.Asteroid{ID: "asteroid-1", X: 30, Y: 40, Size: 2, Variant: 0}
	gameInstance.entities.Pickups["pickup-1"] = &pickups.Pickup{ID: "pickup-1", Type: pickups.TypeOneUp, X: -2, Y: 5}

	control := NewControl(gameInstance)
	bodies := devtools.CollisionBodies(control)
	byKind := make(map[string]devtools.CollisionBody, len(bodies))
	for _, body := range bodies {
		byKind[body.Kind] = body
	}

	player, ok := byKind["player"]
	if !ok {
		t.Fatal("expected player collision body to exist")
	}
	if player.ID != "player-1" {
		t.Fatalf("expected player id %q, got %q", "player-1", player.ID)
	}
	if player.Shape != "rectangle" {
		t.Fatalf("expected player shape %q, got %q", "rectangle", player.Shape)
	}
	if len(player.Points) != 4 {
		t.Fatalf("expected 4 player outline points, got %d", len(player.Points))
	}
	assertCollisionPointApproxEqual(t, player.Points[0], devtools.CollisionPoint{X: 11, Y: 18})

	bullet, ok := byKind["bullet"]
	if !ok { t.Fatal("expected bullet collision body to exist") }
	if bullet.ID != "bullet-1" { t.Fatalf("expected bullet id %q, got %q", "bullet-1", bullet.ID) }
	if bullet.Shape != "circle" { t.Fatalf("expected bullet shape %q, got %q", "circle", bullet.Shape) }
	if len(bullet.Points) != 24 { t.Fatalf("expected %d bullet outline points, got %d", 24, len(bullet.Points)) }
	assertCollisionPointApproxEqual(t, bullet.Points[0], devtools.CollisionPoint{X: 4, Y: 2})

	pickup, ok := byKind["pickup"]
	if !ok { t.Fatal("expected pickup collision body to exist") }
	if pickup.ID != "pickup-1" { t.Fatalf("expected pickup id %q, got %q", "pickup-1", pickup.ID) }
	if pickup.Shape != "circle" { t.Fatalf("expected pickup shape %q, got %q", "circle", pickup.Shape) }
	if len(pickup.Points) != 24 { t.Fatalf("expected %d pickup outline points, got %d", 24, len(pickup.Points)) }
	assertCollisionPointApproxEqual(t, pickup.Points[0], devtools.CollisionPoint{X: 1, Y: 5})

	if _, ok := byKind["asteroid"]; ok { t.Fatalf("expected asteroid collision body to be skipped") }
}

func TestDevtoolsCollisionBodyMarshalsWithLowercaseKeys(t *testing.T) {
	body := devtools.CollisionBody{Kind: "player", ID: "Player-1", Shape: "rectangle", Points: []devtools.CollisionPoint{{X: 1, Y: 2}}}
	data, err := json.Marshal(body)
	if err != nil { t.Fatalf("json.Marshal() error = %v", err) }

	jsonText := string(data)
	assertContains(t, jsonText, `"kind"`)
	assertContains(t, jsonText, `"id"`)
	assertContains(t, jsonText, `"shape"`)
	assertContains(t, jsonText, `"points"`)
	assertContains(t, jsonText, `"x"`)
	assertContains(t, jsonText, `"y"`)
}

func assertCollisionPointApproxEqual(t *testing.T, actual devtools.CollisionPoint, expected devtools.CollisionPoint) {
	t.Helper()
	const epsilon = 1e-9
	if math.Abs(actual.X-expected.X) > epsilon || math.Abs(actual.Y-expected.Y) > epsilon {
		t.Fatalf("expected point %#v, got %#v", expected, actual)
	}
}

func assertContains(t *testing.T, text string, needle string) {
	t.Helper()
	if !strings.Contains(text, needle) { t.Fatalf("expected %q to contain %q", text, needle) }
}
