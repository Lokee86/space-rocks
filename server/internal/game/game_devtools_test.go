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

func TestDebugInfiniteLivesPlayerDiesWithoutLosingLife(t *testing.T) {
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

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugInfiniteLives})
	game.handleShipAsteroidCollisions()

	if !player.IsPendingDespawn() {
		t.Fatal("expected infinite-lives player to still die and despawn")
	}
	if game.playerSessions[playerID].Lives != constants.PlayerStartingLives {
		t.Fatalf("expected infinite-lives player to keep %d lives, got %d", constants.PlayerStartingLives, game.playerSessions[playerID].Lives)
	}
	events := game.pendingEvents[playerID]
	if len(events) != 1 {
		t.Fatalf("expected death event for infinite-lives player, got %d", len(events))
	}
	if events[0].Lives != constants.PlayerStartingLives {
		t.Fatalf("expected death event to keep %d lives, got %d", constants.PlayerStartingLives, events[0].Lives)
	}

	game.Step(constants.CollisionDespawnDelay)
	game.playerSessions[playerID].Step(constants.PlayerRespawnDelay)
	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeRespawn})

	respawned := game.state.Players[playerID]
	if !respawned.DevTools.InfiniteLives {
		t.Fatal("expected infinite lives flag to persist after respawn")
	}
}

func TestDebugInfiniteLivesToggleCanBeDisabled(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	player := game.state.Players[playerID]

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugInfiniteLives})
	if player.DevTools.CanLoseLives() {
		t.Fatal("expected first toggle to enable infinite lives")
	}

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugInfiniteLives})
	if !player.DevTools.CanLoseLives() {
		t.Fatal("expected second toggle to disable infinite lives")
	}
}
