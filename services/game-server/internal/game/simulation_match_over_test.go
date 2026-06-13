package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestStepSkipsAsteroidSpawningAfterMatchOver(t *testing.T) {
	game := newMatchOverTestGame()
	game.cameraViews["player-1"] = &runtime.CameraView{
		X: 100,
		Y: 200,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  640,
			VisibleWorldHeight: 360,
		},
	}

	before := len(game.entities.Asteroids)
	game.Step(constants.AsteroidSpawnInterval + 1)
	after := len(game.entities.Asteroids)

	if after != before {
		t.Fatalf("expected asteroid count to stay at %d after match over, got %d", before, after)
	}
}

func TestStepDoesNotPanicAfterMatchOverWithCleanupSafeEntities(t *testing.T) {
	game := newMatchOverTestGame()
	game.entities.Asteroids["asteroid-1"] = &runtime.Asteroid{
		ID:             "asteroid-1",
		PendingDespawn: true,
		DespawnDelay:   0,
	}
	game.entities.Projectiles["bullet-1"] = &runtime.Bullet{
		ID:             "bullet-1",
		PendingDespawn: true,
		DespawnDelay:   0,
	}
	game.entities.Pickups["pickup-1"] = &pickups.Pickup{
		ID:              "pickup-1",
		Type:            pickups.TypeOneUp,
		LifespanSeconds: 10,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected match-over step to avoid panicking, got %v", r)
		}
	}()

	game.Step(constants.AsteroidSpawnInterval + 1)
}

func newMatchOverTestGame() *Game {
	game := New()
	session := newPlayerSession("player-1", physics.Vector2{X: 100, Y: 200})
	session.Lives = 0
	game.playerSessions[session.ID] = session

	if !game.MatchDecision().IsOver {
		panic("test setup failed: expected game to be match over")
	}

	return game
}
