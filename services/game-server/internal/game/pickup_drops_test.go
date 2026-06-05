package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/drops"
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestMaybeDropPickupFromAsteroidLockedCreatesPickup(t *testing.T) {
	game := New()
	game.dropTables = basicAsteroidsDropTables(2)

	asteroid := &runtime.Asteroid{
		ID:   "asteroid-1",
		Size: 2,
		X:    123,
		Y:    456,
	}

	game.mu.Lock()
	game.maybeDropPickupFromAsteroidLocked(asteroid)
	game.mu.Unlock()

	if len(game.entities.Pickups) != 1 {
		t.Fatalf("expected one pickup, got %d", len(game.entities.Pickups))
	}

	var pickup *pickups.Pickup
	for _, value := range game.entities.Pickups {
		pickup = value
	}
	if pickup == nil {
		t.Fatalf("expected pickup to exist")
	}
	if pickup.Type != "1_up" {
		t.Fatalf("expected pickup type 1_up, got %q", pickup.Type)
	}
	if pickup.X != asteroid.X || pickup.Y != asteroid.Y {
		t.Fatalf("expected pickup position %v,%v, got %v,%v", asteroid.X, asteroid.Y, pickup.X, pickup.Y)
	}
}

func TestMaybeDropPickupFromAsteroidLockedRespectsMaxActivePickups(t *testing.T) {
	game := New()
	game.dropTables = basicAsteroidsDropTables(1)
	game.entities.Pickups["pickup-1"] = &pickups.Pickup{
		ID:   "pickup-1",
		Type: "1_up",
		X:    10,
		Y:    20,
	}

	asteroid := &runtime.Asteroid{
		ID:   "asteroid-1",
		Size: 2,
		X:    123,
		Y:    456,
	}

	game.mu.Lock()
	game.maybeDropPickupFromAsteroidLocked(asteroid)
	game.mu.Unlock()

	if len(game.entities.Pickups) != 1 {
		t.Fatalf("expected pickup count to remain 1, got %d", len(game.entities.Pickups))
	}
	if _, ok := game.entities.Pickups["pickup-1"]; !ok {
		t.Fatalf("expected existing pickup to remain")
	}
}

func TestMaybeDropPickupFromAsteroidLockedDoesNotCreatePickupWhenChanceIsZero(t *testing.T) {
	game := New()
	game.dropTables = basicAsteroidsDropTablesWithChance(0.0)

	asteroid := &runtime.Asteroid{
		ID:   "asteroid-1",
		Size: 2,
		X:    123,
		Y:    456,
	}

	game.mu.Lock()
	game.maybeDropPickupFromAsteroidLocked(asteroid)
	game.mu.Unlock()

	if len(game.entities.Pickups) != 0 {
		t.Fatalf("expected no pickup, got %d", len(game.entities.Pickups))
	}
}

func TestMaybeDropPickupFromAsteroidLockedProjectsPickupIntoStatePacket(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	game.dropTables = basicAsteroidsDropTablesWithChance(1.0)

	asteroid := &runtime.Asteroid{
		ID:   "asteroid-1",
		Size: 2,
		X:    123,
		Y:    456,
	}

	game.mu.Lock()
	game.maybeDropPickupFromAsteroidLocked(asteroid)
	game.mu.Unlock()

	packet := game.StatePacket(playerID)

	if len(packet.Pickups) != 1 {
		t.Fatalf("expected one projected pickup, got %d", len(packet.Pickups))
	}

	var pickup runtime.PickupState
	for _, value := range packet.Pickups {
		pickup = value
	}

	if pickup.Type != "1_up" {
		t.Fatalf("expected pickup type 1_up, got %q", pickup.Type)
	}
	if pickup.X != asteroid.X || pickup.Y != asteroid.Y {
		t.Fatalf("expected pickup position %v,%v, got %v,%v", asteroid.X, asteroid.Y, pickup.X, pickup.Y)
	}
}

func TestApplyProjectileAsteroidHitConsequencesDropsPickup(t *testing.T) {
	game := New()
	game.dropTables = basicAsteroidsDropTables(2)
	asteroid := &runtime.Asteroid{
		ID:   "asteroid-1",
		Size: 2,
		X:    123,
		Y:    456,
	}
	game.entities.Asteroids[asteroid.ID] = asteroid

	game.applyProjectileAsteroidHitConsequences(
		map[string]bool{},
		map[string]*runtime.Asteroid{"asteroid-1": asteroid},
		nil,
	)

	if len(game.entities.Pickups) != 1 {
		t.Fatalf("expected one pickup, got %d", len(game.entities.Pickups))
	}
	if !asteroid.PendingDespawn {
		t.Fatal("expected asteroid to be marked pending despawn")
	}
	if asteroid.DespawnDelay != constants.CollisionDespawnDelay {
		t.Fatalf("expected asteroid despawn delay %v, got %v", constants.CollisionDespawnDelay, asteroid.DespawnDelay)
	}
	var pickup *pickups.Pickup
	for _, value := range game.entities.Pickups {
		pickup = value
	}
	if pickup == nil || pickup.Type != "1_up" {
		t.Fatalf("expected dropped 1_up pickup, got %#v", pickup)
	}
	if pickup.X != asteroid.X || pickup.Y != asteroid.Y {
		t.Fatalf("expected pickup position %v,%v, got %v,%v", asteroid.X, asteroid.Y, pickup.X, pickup.Y)
	}
}

func basicAsteroidsDropTables(maxActivePickups int) drops.Tables {
	return basicAsteroidsDropTablesWithChance(1.0, maxActivePickups)
}

func basicAsteroidsDropTablesWithChance(chance float64, maxActivePickups ...int) drops.Tables {
	activePickups := 2
	if len(maxActivePickups) > 0 {
		activePickups = maxActivePickups[0]
	}
	return drops.Tables{
		ByID: map[string]drops.Table{
			"basicasteroids": {
				ID:                "basicasteroids",
				SourceType:        drops.SourceTypeAsteroid,
				DropMode:          drops.DropModeSingle,
				MaxDropsPerSource: 1,
				MaxActivePickups:  activePickups,
				Entries: []drops.Entry{
					{
						PickupType:    "1_up",
						Chance:        chance,
						MinSourceSize: 1,
						MaxSourceSize: 4,
					},
				},
			},
		},
	}
}
