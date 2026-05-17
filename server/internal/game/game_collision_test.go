package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestHandleBulletAsteroidCollisionsDelaysHitDespawns(t *testing.T) {
	game := New()
	game.collisionShapes = physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{
			Type:   "capsule",
			Radius: 3,
			Height: 24,
		},
		Asteroids: []physics.ImportedCollisionShape{
			{
				Type: "polygon",
				Points: [][]float64{
					{-40, -40},
					{40, -40},
					{40, 40},
					{-40, 40},
				},
			},
		},
	}
	game.state.Projectiles["bullet-1"] = &entities.Bullet{
		ID: "bullet-1",
		X:  100,
		Y:  100,
	}
	game.state.Asteroids["asteroid-1"] = &entities.Asteroid{
		ID:   "asteroid-1",
		X:    100,
		Y:    100,
		Size: 1,
	}
	game.state.Asteroids["asteroid-2"] = &entities.Asteroid{
		ID:   "asteroid-2",
		X:    1000,
		Y:    1000,
		Size: 1,
	}
	game.state.Players["player-1"] = &entities.Ship{ID: "player-1"}
	game.pendingEvents["player-1"] = nil

	game.handleBulletAsteroidCollisions()

	bullet, ok := game.state.Projectiles["bullet-1"]
	if !ok {
		t.Fatal("expected hit bullet to remain during despawn delay")
	}
	if !bullet.PendingDespawn {
		t.Fatal("expected hit bullet to be marked for delayed despawn")
	}

	asteroid, ok := game.state.Asteroids["asteroid-1"]
	if !ok {
		t.Fatal("expected hit asteroid to remain during despawn delay")
	}
	if !asteroid.PendingDespawn {
		t.Fatal("expected hit asteroid to be marked for delayed despawn")
	}
	if _, ok := game.state.Asteroids["asteroid-2"]; !ok {
		t.Fatal("expected untouched asteroid to remain")
	}
	if len(game.pendingEvents["player-1"]) != 1 {
		t.Fatalf("expected 1 queued event, got %d", len(game.pendingEvents["player-1"]))
	}
	if game.pendingEvents["player-1"][0].Type != PacketTypeBulletBlast {
		t.Fatalf("expected bullet_blast event, got %q", game.pendingEvents["player-1"][0].Type)
	}

	packet := game.statePacket("player-1")
	if len(packet.Events) != 1 {
		t.Fatalf("expected 1 event in state packet, got %d", len(packet.Events))
	}

	game.Step(constants.CollisionDespawnDelay)

	if _, ok := game.state.Projectiles["bullet-1"]; ok {
		t.Fatal("expected hit bullet to be removed after despawn delay")
	}
	if _, ok := game.state.Asteroids["asteroid-1"]; ok {
		t.Fatal("expected hit asteroid to be removed after despawn delay")
	}
}

func TestHandleBulletAsteroidCollisionsSplitsLargerAsteroid(t *testing.T) {
	game := New()
	game.collisionShapes = physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{
			Type:   "capsule",
			Radius: 3,
			Height: 24,
		},
		Asteroids: []physics.ImportedCollisionShape{
			{
				Type: "polygon",
				Points: [][]float64{
					{-40, -40},
					{40, -40},
					{40, 40},
					{-40, 40},
				},
			},
		},
	}
	game.state.Projectiles["bullet-1"] = &entities.Bullet{
		ID: "bullet-1",
		X:  100,
		Y:  100,
	}
	game.state.Asteroids["asteroid-1"] = &entities.Asteroid{
		ID:   "asteroid-1",
		X:    100,
		Y:    100,
		Size: 3,
	}

	game.handleBulletAsteroidCollisions()

	if len(game.state.Asteroids) != 3 {
		t.Fatalf("expected hit asteroid plus 2 fragments, got %d asteroids", len(game.state.Asteroids))
	}

	fragmentCount := 0
	for asteroidID, asteroid := range game.state.Asteroids {
		if asteroidID == "asteroid-1" {
			continue
		}

		fragmentCount++
		if asteroid.Size != 2 {
			t.Fatalf("expected fragment size 2, got %d", asteroid.Size)
		}
		if asteroid.X != 100 || asteroid.Y != 100 {
			t.Fatalf("expected fragment at impact position, got (%v, %v)", asteroid.X, asteroid.Y)
		}
		if asteroid.PendingDespawn {
			t.Fatal("expected fragment to remain active")
		}
	}

	if fragmentCount != 2 {
		t.Fatalf("expected 2 fragments, got %d", fragmentCount)
	}
}

func TestHandleShipAsteroidCollisionsDelaysPlayerRemovalAndBroadcastsDeath(t *testing.T) {
	game := New()
	game.collisionShapes = physics.CollisionShapeCatalog{
		Ship: physics.ImportedCollisionShape{
			Type:   "circle",
			Radius: 20,
		},
		Asteroids: []physics.ImportedCollisionShape{
			{
				Type:   "circle",
				Radius: 20,
			},
		},
	}
	game.state.Players["player-1"] = &entities.Ship{
		ID: "player-1",
		X:  100,
		Y:  100,
	}
	game.state.Players["player-2"] = &entities.Ship{
		ID: "player-2",
		X:  1000,
		Y:  1000,
	}
	game.state.Asteroids["asteroid-1"] = &entities.Asteroid{
		ID:   "asteroid-1",
		X:    100,
		Y:    100,
		Size: 1,
	}
	game.pendingEvents["player-1"] = nil
	game.pendingEvents["player-2"] = nil

	game.handleShipAsteroidCollisions()

	player, ok := game.state.Players["player-1"]
	if !ok {
		t.Fatal("expected hit player to remain during despawn delay")
	}
	if !player.PendingDespawn {
		t.Fatal("expected hit player to be marked for delayed despawn")
	}
	if game.state.Asteroids["asteroid-1"].PendingDespawn {
		t.Fatal("expected ship collision to leave asteroid active")
	}
	if _, ok := game.state.Players["player-2"]; !ok {
		t.Fatal("expected untouched player to remain")
	}

	for _, playerID := range []string{"player-1", "player-2"} {
		events := game.pendingEvents[playerID]
		if len(events) != 1 {
			t.Fatalf("expected 1 queued event for %s, got %d", playerID, len(events))
		}
		if events[0].Type != PacketTypeShipDeath {
			t.Fatalf("expected ship_death event for %s, got %q", playerID, events[0].Type)
		}
		if events[0].PlayerID != "player-1" {
			t.Fatalf("expected dead player id player-1 for %s, got %q", playerID, events[0].PlayerID)
		}
		if events[0].X != 100 || events[0].Y != 100 {
			t.Fatalf("expected death event at player position for %s, got (%v, %v)", playerID, events[0].X, events[0].Y)
		}
	}

	packet := game.statePacket("player-1")
	if len(packet.Events) != 1 {
		t.Fatalf("expected hit player to receive death event, got %d events", len(packet.Events))
	}

	game.Step(constants.CollisionDespawnDelay)

	if _, ok := game.state.Players["player-1"]; ok {
		t.Fatal("expected hit player to be removed after despawn delay")
	}
	if _, ok := game.state.Players["player-2"]; !ok {
		t.Fatal("expected untouched player to remain after despawn delay")
	}
}

func TestAsteroidVisibilityUsesCameraViewsWithoutPlayer(t *testing.T) {
	game := New()
	game.cameraViews["player-1"] = &entities.CameraView{
		X: 100,
		Y: 100,
		Config: entities.ClientConfig{
			VisibleWorldWidth:  200,
			VisibleWorldHeight: 200,
		},
	}
	asteroid := &entities.Asteroid{
		ID: "asteroid-1",
		X:  100,
		Y:  100,
	}

	if game.isAsteroidFarFromAllCameras(asteroid) {
		t.Fatal("expected asteroid inside camera view to remain even without a player entity")
	}
}

func TestStateFlushesEventsForPlayer(t *testing.T) {
	game := New()
	game.state.Players["player-1"] = &entities.Ship{ID: "player-1"}
	game.pendingEvents["player-1"] = []EventState{{Type: PacketTypeBulletBlast, X: 10, Y: 20}}

	first := game.State("player-1")
	if first == nil {
		t.Fatal("expected first state response")
	}
	if len(game.pendingEvents["player-1"]) != 0 {
		t.Fatal("expected State to flush queued events")
	}

	packet := game.statePacket("player-1")
	if len(packet.Events) != 0 {
		t.Fatalf("expected flushed state packet to have 0 events, got %d", len(packet.Events))
	}
}
