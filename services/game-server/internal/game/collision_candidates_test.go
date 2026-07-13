package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func candidateBody(x, y float64) physics.CollisionBody {
	return physics.CollisionBody{
		Position: physics.Vector2{X: x, Y: y},
		Shape:    physics.NewCircleShape(5),
	}
}

func TestCollisionOuterIDsAreSortedAndRefilled(t *testing.T) {
	game := spatialTestGame()
	game.entities.Players["Player-2"] = &runtime.Ship{}
	game.entities.Players["Player-1"] = &runtime.Ship{}
	game.entities.Players["stale"] = nil
	game.entities.Projectiles["Projectile-2"] = &runtime.Bullet{}
	game.entities.Projectiles["Projectile-1"] = &runtime.Bullet{}

	players := game.collisionPlayerIDsSorted()
	projectiles := game.collisionProjectileIDsSorted()
	if got := len(players); got != 2 || players[0] != "Player-1" || players[1] != "Player-2" {
		t.Fatalf("players = %#v", players)
	}
	if got := len(projectiles); got != 2 || projectiles[0] != "Projectile-1" || projectiles[1] != "Projectile-2" {
		t.Fatalf("projectiles = %#v", projectiles)
	}
	delete(game.entities.Players, "Player-1")
	if got := game.collisionPlayerIDsSorted(); len(got) != 1 || got[0] != "Player-2" {
		t.Fatalf("refilled players = %#v", got)
	}
}

func TestAsteroidCandidatesNearestWrappedThenID(t *testing.T) {
	game := spatialTestGame()
	game.entities.Asteroids["far"] = &runtime.Asteroid{ID: "far", X: 5, Y: 0, Size: 1}
	game.entities.Asteroids["edge"] = &runtime.Asteroid{ID: "edge", X: constants.WorldWidth - 2, Y: 0, Size: 1}
	game.entities.Asteroids["tie-b"] = &runtime.Asteroid{ID: "tie-b", X: 2, Y: 0, Size: 1}
	game.entities.Asteroids["tie-a"] = &runtime.Asteroid{ID: "tie-a", X: 2, Y: 0, Size: 1}
	game.rebuildAsteroidSpatialIndex()

	refs := game.asteroidCollisionCandidates(candidateBody(0, 0))
	if len(refs) < 4 || refs[0].ID != "edge" || refs[1].ID != "tie-a" || refs[2].ID != "tie-b" || refs[3].ID != "far" {
		t.Fatalf("asteroid candidates = %#v", refs)
	}
}

func TestPickupCandidatesNearestAndStaleFiltered(t *testing.T) {
	game := spatialTestGame()
	game.entities.Pickups["far"] = pickupForTest("far", 5, 0)
	game.entities.Pickups["near"] = pickupForTest("near", 1, 0)
	game.rebuildPickupSpatialIndex()
	delete(game.entities.Pickups, "near")

	refs := game.pickupCollisionCandidates(candidateBody(0, 0))
	if len(refs) != 1 || refs[0].ID != "far" {
		t.Fatalf("pickup candidates = %#v", refs)
	}
}

func pickupForTest(id string, x, y float64) *pickups.Pickup {
	return &pickups.Pickup{ID: id, Type: pickups.TypeOneUp, X: x, Y: y}
}


