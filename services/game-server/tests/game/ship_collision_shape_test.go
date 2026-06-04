package gametests

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDefaultShipCollisionShapeIDResolves(t *testing.T) {
	catalog := testShipCollisionCatalog()

	shape, err := catalog.ShipShapeByID(runtime.DefaultShipStats().CollisionShapeID)
	if err != nil {
		t.Fatal(err)
	}

	assertCircleShape(t, shape, 20)
}

func TestDefaultShipCollisionBodyMatchesDefaultShape(t *testing.T) {
	catalog := testShipCollisionCatalog()
	ship := runtime.Ship{
		ID:       "player-1",
		Stats:    runtime.DefaultShipStats(),
		X:        10,
		Y:        20,
		Rotation: 1.5,
	}

	body, ok := ship.CollisionBody(catalog)
	if !ok {
		t.Fatal("expected default ship collision body")
	}
	defaultShape, err := catalog.ShipShape()
	if err != nil {
		t.Fatal(err)
	}

	if body.ID != ship.ID {
		t.Fatalf("expected collision body ID %q, got %q", ship.ID, body.ID)
	}
	if body.Position.X != ship.X || body.Position.Y != ship.Y {
		t.Fatalf("expected collision body position (%v, %v), got (%v, %v)", ship.X, ship.Y, body.Position.X, body.Position.Y)
	}
	if body.Rotation != ship.Rotation {
		t.Fatalf("expected collision body rotation %v, got %v", ship.Rotation, body.Rotation)
	}
	assertSameCircleShape(t, body.Shape, defaultShape)
}

func TestShipCollisionBodyFallsBackForUnknownCollisionShapeID(t *testing.T) {
	catalog := testShipCollisionCatalog()
	ship := runtime.Ship{
		ID: "player-1",
		Stats: runtime.ShipStats{
			CollisionShapeID: "unknown_ship",
		},
	}

	body, ok := ship.CollisionBody(catalog)
	if !ok {
		t.Fatal("expected unknown ship collision shape ID to fall back")
	}

	assertCircleShape(t, body.Shape, 20)
}

func TestRespawnSafetyFallsBackForUnknownSessionCollisionShapeID(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.removePlayerEntity(playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{}, 1)
	stats := runtime.DefaultShipStats()
	stats.CollisionShapeID = "unknown_ship"
	scenario.sessionField(playerID, "Stats").Set(reflect.ValueOf(stats))

	insideBuffer := physics.Vector2{X: constants.PlayerRespawnBuffer + 21, Y: 0}
	scenario.setSessionSpawnPosition(playerID, insideBuffer)
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	respawned := scenario.playerState(playerID, playerID)
	if respawned.X == insideBuffer.X && respawned.Y == insideBuffer.Y {
		t.Fatal("expected unknown session collision shape ID to fall back and avoid unsafe respawn")
	}
}

func testShipCollisionCatalog() physics.CollisionShapeCatalog {
	return physics.CollisionShapeCatalog{
		Ship: physics.ImportedCollisionShape{
			Type:   "circle",
			Radius: 20,
		},
	}
}

func assertCircleShape(t *testing.T, shape physics.CollisionShape, radius float64) {
	t.Helper()

	if shape.Type != physics.CollisionShapeCircle || shape.Radius != radius {
		t.Fatalf("expected circle shape with radius %v, got %#v", radius, shape)
	}
}

func assertSameCircleShape(t *testing.T, shape physics.CollisionShape, expected physics.CollisionShape) {
	t.Helper()

	if shape.Type != expected.Type || shape.Radius != expected.Radius {
		t.Fatalf("expected circle shape %#v, got %#v", expected, shape)
	}
}
