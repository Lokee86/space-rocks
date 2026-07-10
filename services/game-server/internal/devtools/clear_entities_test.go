package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/devtools/streamruntime"
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

func TestHandleDebugClearBulletsUsesInjectedStreamRuntime(t *testing.T) {
	injectedRuntime := streamruntime.NewRuntime()
	streamruntime.DefaultRuntime.ClearContinuousBulletStreams()
	t.Cleanup(func() {
		streamruntime.DefaultRuntime.ClearContinuousBulletStreams()
	})

	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := gameInstance.AddPlayer()

	bulletA := runtime.NewBullet("debug-bullet-a", playerID, physics.Vector2{X: 10, Y: 20}, 0, physics.Vector2{}, 5)
	bulletB := runtime.NewBullet("debug-bullet-b", playerID, physics.Vector2{X: 30, Y: 40}, 0, physics.Vector2{}, 5)
	control.AddBullet(bulletA)
	control.AddBullet(bulletB)

	if !injectedRuntime.BeginContinuousBulletStream(playerID, physics.Vector2{X: 10, Y: 20}, physics.Vector2{X: 0, Y: -1}) {
		t.Fatal("expected injected runtime stream to start")
	}
	if !streamruntime.DefaultRuntime.BeginContinuousBulletStream(playerID, physics.Vector2{X: 20, Y: 30}, physics.Vector2{X: 1, Y: 0}) {
		t.Fatal("expected default runtime stream to start")
	}

	controller := NewController(Dependencies{Target: control, Streams: injectedRuntime})
	if !controller.HandleCommand(playerID, DebugCommand{Type: PacketTypeDebugClearBullets}) {
		t.Fatalf("expected HandleCommand to return true")
	}

	snapshot := gameInstance.GameplayPresentationSnapshot(playerID)
	if len(snapshot.Bullets) != 0 {
		t.Fatalf("expected 0 bullets after clear, got %d", len(snapshot.Bullets))
	}
	if got := len(injectedRuntime.ActiveContinuousBulletStreams()); got != 0 {
		t.Fatalf("expected injected runtime to be cleared, got %d active streams", got)
	}
	if got := len(streamruntime.DefaultRuntime.ActiveContinuousBulletStreams()); got != 1 {
		t.Fatalf("expected default runtime to remain uncleared, got %d active streams", got)
	}
}

func TestHandleDebugClearAsteroidsRemovesAllAsteroids(t *testing.T) {
	gameInstance := game.New()
	control := game.NewControl(gameInstance)
	playerID := gameInstance.AddPlayer()
	gameInstance.SetPlayerScore(playerID, 25)

	control.ApplyAsteroidSpawnPlan(spawning.AsteroidSpawnPlan{
		EntityType: spawning.SpawnEntityTypeAsteroid,
		Reason:     spawning.SpawnReasonDebugAsteroid,
		Position:   physics.Vector2{X: 100, Y: 100},
		Velocity:   physics.Vector2{X: 1, Y: 0},
		Size:       3,
		Variant:    0,
	})
	control.ApplyAsteroidSpawnPlan(spawning.AsteroidSpawnPlan{
		EntityType: spawning.SpawnEntityTypeAsteroid,
		Reason:     spawning.SpawnReasonDebugAsteroid,
		Position:   physics.Vector2{X: 200, Y: 200},
		Velocity:   physics.Vector2{X: -1, Y: 0},
		Size:       2,
		Variant:    1,
	})

	if !HandleCommand(control, playerID, DebugCommand{Type: PacketTypeDebugClearAsteroids}) {
		t.Fatalf("expected HandleCommand to return true")
	}

	snapshot := gameInstance.GameplayPresentationSnapshot(playerID)
	if len(snapshot.Asteroids) != 0 {
		t.Fatalf("expected 0 asteroids after clear, got %d", len(snapshot.Asteroids))
	}
	session, exists := snapshot.PlayerSessions[playerID]
	if !exists {
		t.Fatalf("expected player session %q in gameplay snapshot", playerID)
	}
	if session.Score != 25 {
		t.Fatalf("expected player score to remain 25, got %d", session.Score)
	}
}
