package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDebugInvinciblePlayerDoesNotDieFromAsteroidCollision(t *testing.T) {
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
	playerID := game.AddPlayer()
	player := game.state.Players[playerID]
	game.state.Asteroids["asteroid-1"] = &entities.Asteroid{
		ID:   "asteroid-1",
		X:    player.X,
		Y:    player.Y,
		Size: 1,
	}

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugInvincible})
	game.handleShipAsteroidCollisions()

	if player.IsPendingDespawn() {
		t.Fatal("expected invincible player not to be marked for despawn")
	}
	if game.playerSessions[playerID].Lives != constants.PlayerStartingLives {
		t.Fatalf("expected invincible player to keep %d lives, got %d", constants.PlayerStartingLives, game.playerSessions[playerID].Lives)
	}
	if len(game.pendingEvents[playerID]) != 0 {
		t.Fatalf("expected no death events for invincible player, got %d", len(game.pendingEvents[playerID]))
	}
}

func TestDebugInvincibleToggleCanBeDisabled(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.state.Players[playerID]

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugInvincible})
	if player.DevTools.CanTakeDamage() {
		t.Fatal("expected first toggle to make player invincible")
	}

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugInvincible})
	if !player.DevTools.CanTakeDamage() {
		t.Fatal("expected second toggle to make player vulnerable")
	}
}
