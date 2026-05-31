package devtools

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/spawning"
)

func TestHandleDebugClearBulletsRemovesAllBullets(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	bulletA := entities.NewBullet("debug-bullet-a", playerID, physics.Vector2{X: 10, Y: 20}, 0, physics.Vector2{}, 5)
	bulletB := entities.NewBullet("debug-bullet-b", playerID, physics.Vector2{X: 30, Y: 40}, 0, physics.Vector2{}, 5)
	target.DevtoolsAddBullet(bulletA)
	target.DevtoolsAddBullet(bulletB)

	ok := HandleCommand(target, playerID, DebugCommand{Type: PacketTypeDebugClearBullets})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	packet := target.StatePacket(playerID)
	if len(packet.Bullets) != 0 {
		t.Fatalf("expected 0 bullets after clear, got %d", len(packet.Bullets))
	}
}

func TestHandleDebugClearBulletsIsSafeWhenEmpty(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	ok := HandleCommand(target, playerID, DebugCommand{Type: PacketTypeDebugClearBullets})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	packet := target.StatePacket(playerID)
	if len(packet.Bullets) != 0 {
		t.Fatalf("expected 0 bullets after clear, got %d", len(packet.Bullets))
	}
}

func TestHandleDebugClearAsteroidsRemovesAllAsteroids(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()
	target.SetPlayerScore(playerID, 25)

	target.DevtoolsApplyAsteroidSpawnPlan(spawning.AsteroidSpawnPlan{
		EntityType: spawning.SpawnEntityTypeAsteroid,
		Reason:     spawning.SpawnReasonDebugAsteroid,
		Position:   physics.Vector2{X: 100, Y: 100},
		Velocity:   physics.Vector2{X: 1, Y: 0},
		Size:       3,
		Variant:    0,
	})
	target.DevtoolsApplyAsteroidSpawnPlan(spawning.AsteroidSpawnPlan{
		EntityType: spawning.SpawnEntityTypeAsteroid,
		Reason:     spawning.SpawnReasonDebugAsteroid,
		Position:   physics.Vector2{X: 200, Y: 200},
		Velocity:   physics.Vector2{X: -1, Y: 0},
		Size:       2,
		Variant:    1,
	})

	ok := HandleCommand(target, playerID, DebugCommand{Type: PacketTypeDebugClearAsteroids})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	packet := target.StatePacket(playerID)
	if len(packet.Asteroids) != 0 {
		t.Fatalf("expected 0 asteroids after clear, got %d", len(packet.Asteroids))
	}
	player, exists := packet.Players[playerID]
	if !exists {
		t.Fatalf("expected player %q in state packet", playerID)
	}
	if player.Score != 25 {
		t.Fatalf("expected player score to remain 25, got %d", player.Score)
	}
}

func TestHandleDebugClearAsteroidsIsSafeWhenEmpty(t *testing.T) {
	target := game.New()
	playerID := target.AddPlayer()

	ok := HandleCommand(target, playerID, DebugCommand{Type: PacketTypeDebugClearAsteroids})
	if !ok {
		t.Fatalf("expected HandleCommand to return true")
	}

	packet := target.StatePacket(playerID)
	if len(packet.Asteroids) != 0 {
		t.Fatalf("expected 0 asteroids after clear, got %d", len(packet.Asteroids))
	}
}
