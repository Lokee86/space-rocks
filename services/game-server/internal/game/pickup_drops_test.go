package game

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/drops"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
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

func TestMaybeDropPickupFromAsteroidLockedIsDeterministicForSeed(t *testing.T) {
	const seed int64 = 8675309

	gameA := NewWithSeed(seed)
	gameB := NewWithSeed(seed)
	gameA.dropTables = basicAsteroidsDropTablesWithChance(0.5, 4)
	gameB.dropTables = basicAsteroidsDropTablesWithChance(0.5, 4)

	asteroids := []*runtime.Asteroid{
		{ID: "asteroid-1", Size: 2, X: 123, Y: 456},
		{ID: "asteroid-2", Size: 3, X: 234, Y: 567},
		{ID: "asteroid-3", Size: 4, X: 345, Y: 678},
	}

	for index, asteroid := range asteroids {
		gameA.mu.Lock()
		gameA.maybeDropPickupFromAsteroidLocked(asteroid)
		gameA.mu.Unlock()

		gameB.mu.Lock()
		gameB.maybeDropPickupFromAsteroidLocked(asteroid)
		gameB.mu.Unlock()

		if !reflect.DeepEqual(gameA.entities.Pickups, gameB.entities.Pickups) {
			t.Fatalf("after asteroid %d: game A pickups %#v, game B pickups %#v", index, gameA.entities.Pickups, gameB.entities.Pickups)
		}
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
	game.rebuildPickupSpatialIndex()
	refs := game.spatialIndex.QueryCircle(nil, physics.Vector2{X: asteroid.X, Y: asteroid.Y}, 0, spatial.KindMask(spatial.KindPickup))
	if len(refs) != 1 || refs[0].ID != pickup.ID {
		t.Fatalf("expected same-tick pickup index entry, got %#v", refs)
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
