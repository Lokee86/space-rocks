package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/rng"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial/grid"
)

func spatialTestGame() *Game {
	lifeRuntime, err := lives.NewRuntime(lives.NewBaselinePolicy())
	if err != nil {
		panic(err)
	}
	return &Game{
		rngSource: rng.New(1),
		collisionShapes: physics.CollisionShapeCatalog{
			Asteroids: []physics.ImportedCollisionShape{{Type: "circle", Radius: 4}},
			Pickups: map[string]physics.ImportedCollisionShape{
				constants.PickupOneUpClass: {Type: "circle", Radius: 2},
			},
		},
		entities:     runtime.NewEntityStore(),
		lifeRuntime:  lifeRuntime,
		spatialIndex: grid.New(space.DefaultBounds(), defaultSpatialCellSize),
	}
}

func TestNewOwnsSpatialIndex(t *testing.T) {
	game := New()
	if game.spatialIndex == nil {
		t.Fatal("New() did not create a spatial index")
	}
}

func TestRebuildAsteroidSpatialIndexProjectsActiveEntries(t *testing.T) {
	game := spatialTestGame()
	game.entities.Asteroids["active"] = &runtime.Asteroid{ID: "active", X: 10, Y: 20, Size: 1}
	game.entities.Asteroids["pending"] = &runtime.Asteroid{ID: "pending", PendingDespawn: true, Size: 1}
	body, ok := game.entities.Asteroids["active"].CollisionBody(game.collisionShapes)
	if !ok {
		t.Fatal("active asteroid collision body was not created")
	}
	expectedRadius := physics.BoundingRadius(body.Shape)

	game.rebuildAsteroidSpatialIndex()
	refs := game.spatialIndex.QueryCircle(nil, physics.Vector2{X: 10, Y: 20}, 0, spatial.AllKinds)
	if len(refs) != 1 || refs[0].Kind != spatial.KindAsteroid || refs[0].ID != "active" {
		t.Fatalf("asteroid refs = %#v", refs)
	}
	if len(game.spatialEntries) != 1 || game.spatialEntries[0].Radius != expectedRadius {
		t.Fatalf("projected entries = %#v", game.spatialEntries)
	}
}

func TestRebuildPickupSpatialIndexReplacesAsteroids(t *testing.T) {
	game := spatialTestGame()
	game.entities.Asteroids["asteroid"] = &runtime.Asteroid{ID: "asteroid", Size: 1}
	game.rebuildAsteroidSpatialIndex()
	game.entities.Pickups["pickup"] = &pickups.Pickup{ID: "pickup", Type: pickups.TypeOneUp, X: 30, Y: 40}
	game.rebuildPickupSpatialIndex()

	refs := game.spatialIndex.QueryCircle(nil, physics.Vector2{X: 30, Y: 40}, 0, spatial.AllKinds)
	if len(refs) != 1 || refs[0].Kind != spatial.KindPickup || refs[0].ID != "pickup" {
		t.Fatalf("pickup refs = %#v", refs)
	}
}

func TestRebuildSpatialIndexSkipsMissingShapes(t *testing.T) {
	game := spatialTestGame()
	game.collisionShapes.Asteroids = nil
	game.entities.Asteroids["missing"] = &runtime.Asteroid{ID: "missing", Size: 1}
	game.rebuildAsteroidSpatialIndex()
	if len(game.spatialEntries) != 0 {
		t.Fatalf("missing-shape entries = %#v, want none", game.spatialEntries)
	}
}
