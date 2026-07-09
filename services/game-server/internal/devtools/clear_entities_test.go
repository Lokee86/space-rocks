package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/devtools/streamruntime"
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

func TestHandleDebugClearBulletsRemovesAllBullets(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	playerID := target.AddPlayer()

	bulletA := runtime.NewBullet("debug-bullet-a", playerID, physics.Vector2{X: 10, Y: 20}, 0, physics.Vector2{}, 5)
	bulletB := runtime.NewBullet("debug-bullet-b", playerID, physics.Vector2{X: 30, Y: 40}, 0, physics.Vector2{}, 5)
	control.AddBullet(bulletA)
	control.AddBullet(bulletB)

	ok := HandleCommand(control, playerID, DebugCommand{Type: PacketTypeDebugClearBullets})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	snapshot := target.GameplayPresentationSnapshot(playerID)
	if len(snapshot.Bullets) != 0 {
		t.Fatalf("expected 0 bullets after clear, got %d", len(snapshot.Bullets))
	}
}

func TestHandleDebugClearBulletsClearsContinuousBulletStreams(t *testing.T) {
	streamruntime.DefaultRuntime.ClearContinuousBulletStreams()
	t.Cleanup(func() {
		streamruntime.DefaultRuntime.ClearContinuousBulletStreams()
	})

	target := game.New()
	control := game.NewControl(target)
	playerID := target.AddPlayer()

	if !streamruntime.DefaultRuntime.BeginContinuousBulletStream(playerID, physics.Vector2{X: 10, Y: 20}, physics.Vector2{X: 0, Y: -1}) {
		t.Fatalf("expected to begin continuous bullet stream")
	}

	ok := HandleCommand(control, playerID, DebugCommand{Type: PacketTypeDebugClearBullets})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	if got := len(streamruntime.DefaultRuntime.ActiveContinuousBulletStreams()); got != 0 {
		t.Fatalf("expected 0 active continuous bullet streams after clear, got %d", got)
	}
}

func TestHandleDebugClearBulletsIsSafeWhenEmpty(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	playerID := target.AddPlayer()

	ok := HandleCommand(control, playerID, DebugCommand{Type: PacketTypeDebugClearBullets})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	snapshot := target.GameplayPresentationSnapshot(playerID)
	if len(snapshot.Bullets) != 0 {
		t.Fatalf("expected 0 bullets after clear, got %d", len(snapshot.Bullets))
	}
}

func TestHandleDebugClearAsteroidsRemovesAllAsteroids(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	playerID := target.AddPlayer()
	target.SetPlayerScore(playerID, 25)

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

	ok := HandleCommand(control, playerID, DebugCommand{Type: PacketTypeDebugClearAsteroids})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	snapshot := target.GameplayPresentationSnapshot(playerID)
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

func TestHandleDebugClearAsteroidsIsSafeWhenEmpty(t *testing.T) {
	target := game.New()
	control := game.NewControl(target)
	playerID := target.AddPlayer()

	ok := HandleCommand(control, playerID, DebugCommand{Type: PacketTypeDebugClearAsteroids})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	snapshot := target.GameplayPresentationSnapshot(playerID)
	if len(snapshot.Asteroids) != 0 {
		t.Fatalf("expected 0 asteroids after clear, got %d", len(snapshot.Asteroids))
	}
}
