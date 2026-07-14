package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rng"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spawning"
)

func integrationGame() *Game {
	game := spatialTestGame()
	game.spawner = spawning.New(rng.New(1))
	game.collisionShapes.Ship = physics.ImportedCollisionShape{Type: "circle", Radius: 2}
	game.collisionShapes.Bullet = physics.ImportedCollisionShape{Type: "circle", Radius: 1}
	return game
}

func TestShipAsteroidCandidatesChooseNearestOverlappingAsteroid(t *testing.T) {
	game := integrationGame()
	player := &runtime.Ship{ID: "player", X: 0, Y: 0}
	game.entities.Players[player.ID] = player
	game.entities.Asteroids["far"] = &runtime.Asteroid{ID: "far", X: 5, Y: 0, Size: 1}
	game.entities.Asteroids["near"] = &runtime.Asteroid{ID: "near", X: 3, Y: 0, Size: 1}
	game.rebuildAsteroidSpatialIndex()
	body, ok := player.CollisionBody(game.collisionShapes)
	if !ok {
		t.Fatal("player collision body did not resolve")
	}
	candidates := game.asteroidCollisionCandidates(body)
	if len(candidates) == 0 || candidates[0].ID != "near" {
		t.Fatalf("candidates = %#v", candidates)
	}
	asteroid := game.entities.Asteroids[candidates[0].ID]
	if _, ok := detectPlayerAsteroidCollision(player.ID, player, asteroid, game.collisionShapes); !ok {
		t.Fatal("nearest candidate did not overlap")
	}
}

func TestShipAsteroidCollisionCandidateCrossesWorldBoundary(t *testing.T) {
	game := integrationGame()
	player := &runtime.Ship{ID: "player", X: 0, Y: 0}
	game.entities.Players[player.ID] = player
	game.entities.Asteroids["edge"] = &runtime.Asteroid{ID: "edge", X: constants.WorldWidth - 1, Y: 0, Size: 1}
	game.rebuildAsteroidSpatialIndex()
	body, _ := player.CollisionBody(game.collisionShapes)
	candidates := game.asteroidCollisionCandidates(body)
	if len(candidates) != 1 || candidates[0].ID != "edge" {
		t.Fatalf("candidates = %#v", candidates)
	}
	if _, ok := detectPlayerAsteroidCollision(player.ID, player, game.entities.Asteroids["edge"], game.collisionShapes); !ok {
		t.Fatal("boundary collision did not resolve")
	}
}

func TestUnresolvedPlayerCollisionShapeProducesNoCollision(t *testing.T) {
	game := spatialTestGame()
	player := &runtime.Ship{ID: "player", X: 0, Y: 0}
	game.entities.Players[player.ID] = player
	game.entities.Asteroids["asteroid"] = &runtime.Asteroid{ID: "asteroid", Size: 1}
	game.rebuildAsteroidSpatialIndex()
	_, ok := player.CollisionBody(game.collisionShapes)
	if ok {
		t.Fatal("unresolved player collision body unexpectedly resolved")
	}
}

func TestProjectileAsteroidCandidatesChooseNearestOverlappingAsteroid(t *testing.T) {
	game := integrationGame()
	bullet := &runtime.Bullet{ID: "bullet", X: 0, Y: 0}
	game.entities.Projectiles[bullet.ID] = bullet
	game.entities.Asteroids["far"] = &runtime.Asteroid{ID: "far", X: 2.3, Y: 0, Size: 1}
	game.entities.Asteroids["near"] = &runtime.Asteroid{ID: "near", X: 1.5, Y: 0, Size: 1}
	game.rebuildAsteroidSpatialIndex()
	body, ok := bullet.CollisionBody(game.collisionShapes)
	if !ok {
		t.Fatal("projectile collision body did not resolve")
	}
	candidates := game.asteroidCollisionCandidates(body)
	if len(candidates) == 0 || candidates[0].ID != "near" {
		t.Fatalf("candidates = %#v", candidates)
	}
	if _, ok := detectProjectileAsteroidCollision(bullet, game.entities.Asteroids[candidates[0].ID], game.collisionShapes); !ok {
		t.Fatal("nearest candidate did not overlap")
	}
}

func TestProjectileIterationIsDeterministicForContendingAsteroid(t *testing.T) {
	game := integrationGame()
	game.entities.Projectiles["bullet-b"] = &runtime.Bullet{ID: "bullet-b", X: 0, Y: 0}
	game.entities.Projectiles["bullet-a"] = &runtime.Bullet{ID: "bullet-a", X: 0, Y: 0}
	ids := game.collisionProjectileIDsSorted()
	if len(ids) != 2 || ids[0] != "bullet-a" || ids[1] != "bullet-b" {
		t.Fatalf("projectile order = %#v", ids)
	}
}

func TestProjectileAsteroidCollisionCandidateCrossesWorldBoundary(t *testing.T) {
	game := integrationGame()
	bullet := &runtime.Bullet{ID: "bullet", X: 0, Y: 0}
	game.entities.Projectiles[bullet.ID] = bullet
	game.entities.Asteroids["edge"] = &runtime.Asteroid{ID: "edge", X: constants.WorldWidth - 1, Y: 0, Size: 1}
	game.rebuildAsteroidSpatialIndex()
	body, _ := bullet.CollisionBody(game.collisionShapes)
	candidates := game.asteroidCollisionCandidates(body)
	if len(candidates) != 1 || candidates[0].ID != "edge" {
		t.Fatalf("candidates = %#v", candidates)
	}
	if _, ok := detectProjectileAsteroidCollision(bullet, game.entities.Asteroids["edge"], game.collisionShapes); !ok {
		t.Fatal("boundary collision did not resolve")
	}
}

func TestUnresolvedProjectileCollisionShapeProducesNoCollision(t *testing.T) {
	game := spatialTestGame()
	bullet := &runtime.Bullet{ID: "bullet", X: 0, Y: 0}
	_, ok := bullet.CollisionBody(game.collisionShapes)
	if ok {
		t.Fatal("unresolved projectile collision body unexpectedly resolved")
	}
}

func TestPlayerPickupCandidatesChooseNearestAndWrappedPickup(t *testing.T) {
	game := integrationGame()
	player := &runtime.Ship{ID: "player", X: 0, Y: 0}
	game.entities.Players[player.ID] = player
	game.entities.Pickups["far"] = &pickups.Pickup{ID: "far", Type: pickups.TypeOneUp, X: 5, Y: 0}
	game.entities.Pickups["near"] = &pickups.Pickup{ID: "near", Type: pickups.TypeOneUp, X: 0.5, Y: 0}
	game.entities.Pickups["edge"] = &pickups.Pickup{ID: "edge", Type: pickups.TypeOneUp, X: constants.WorldWidth - 1, Y: 0}
	game.rebuildPickupSpatialIndex()
	body, _ := player.CollisionBody(game.collisionShapes)
	candidates := game.pickupCollisionCandidates(body)
	if len(candidates) < 2 || candidates[0].ID != "near" || candidates[1].ID != "edge" {
		t.Fatalf("candidates = %#v", candidates)
	}
}

func TestPlayerPickupContentionUsesDeterministicPlayerOrder(t *testing.T) {
	game := integrationGame()
	game.entities.Players["player-b"] = &runtime.Ship{ID: "player-b", X: 0, Y: 0}
	game.entities.Players["player-a"] = &runtime.Ship{ID: "player-a", X: 0, Y: 0}
	ids := game.collisionPlayerIDsSorted()
	if len(ids) != 2 || ids[0] != "player-a" || ids[1] != "player-b" {
		t.Fatalf("player order = %#v", ids)
	}
}

func TestUnresolvedPlayerShapeProducesNoPickupCollision(t *testing.T) {
	game := spatialTestGame()
	player := &runtime.Ship{ID: "player", X: 0, Y: 0}
	game.entities.Players[player.ID] = player
	game.entities.Pickups["pickup"] = &pickups.Pickup{ID: "pickup", Type: pickups.TypeOneUp, X: 0, Y: 0}
	_, ok := player.CollisionBody(game.collisionShapes)
	if ok {
		t.Fatal("unresolved player collision body unexpectedly resolved")
	}
}

func TestPickupDroppedByAsteroidConsequencesIsAvailableThisCollisionPhase(t *testing.T) {
	game := integrationGame()
	game.dropTables = basicAsteroidsDropTables(2)
	player := &runtime.Ship{ID: "player", X: 123, Y: 456}
	game.entities.Players[player.ID] = player
	asteroid := &runtime.Asteroid{ID: "asteroid", Size: 2, X: 123, Y: 456}
	game.entities.Asteroids[asteroid.ID] = asteroid
	game.applyProjectileAsteroidHitConsequences(map[string]bool{}, map[string]*runtime.Asteroid{asteroid.ID: asteroid}, nil)
	game.rebuildPickupSpatialIndex()
	game.handlePlayerPickupCollisions()
	if len(game.entities.Pickups) != 0 {
		t.Fatalf("same-tick pickup was not collected: %#v", game.entities.Pickups)
	}
}
