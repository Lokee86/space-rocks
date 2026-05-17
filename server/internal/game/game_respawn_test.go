package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestPlayerDeathReducesLivesAndAllowsRespawnAfterDelay(t *testing.T) {
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
	spawnPosition := player.Position()
	game.state.Asteroids["asteroid-1"] = &entities.Asteroid{
		ID:   "asteroid-1",
		X:    player.X,
		Y:    player.Y,
		Size: 1,
	}

	game.handleShipAsteroidCollisions()

	session := game.playerSessions[playerID]
	if session.Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected %d lives after death, got %d", constants.PlayerStartingLives-1, session.Lives)
	}
	events := game.pendingEvents[playerID]
	if len(events) != 1 {
		t.Fatalf("expected 1 death event, got %d", len(events))
	}
	if events[0].Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected death event lives %d, got %d", constants.PlayerStartingLives-1, events[0].Lives)
	}
	if events[0].RespawnDelay != constants.PlayerRespawnDelay {
		t.Fatalf("expected respawn delay %v, got %v", constants.PlayerRespawnDelay, events[0].RespawnDelay)
	}
	if packet := game.statePacket(playerID); packet.Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected state packet lives %d after death, got %d", constants.PlayerStartingLives-1, packet.Lives)
	}

	game.Step(constants.CollisionDespawnDelay)
	if _, ok := game.state.Players[playerID]; ok {
		t.Fatal("expected dead player entity to be removed before respawn")
	}
	if packet := game.statePacket(playerID); packet.Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected state packet lives %d after player removal, got %d", constants.PlayerStartingLives-1, packet.Lives)
	}

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeRespawn})
	if _, ok := game.state.Players[playerID]; ok {
		t.Fatal("expected respawn to be blocked before delay finishes")
	}

	session.Step(constants.PlayerRespawnDelay)
	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeRespawn})

	respawned, ok := game.state.Players[playerID]
	if !ok {
		t.Fatal("expected player to respawn after delay")
	}
	if respawned.Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected respawned player to keep %d lives, got %d", constants.PlayerStartingLives-1, respawned.Lives)
	}
	if respawned.X == spawnPosition.X && respawned.Y == spawnPosition.Y {
		t.Fatal("expected respawn to avoid asteroid on original spawn point")
	}
	if !game.isSafeRespawnPosition(respawned.Position()) {
		t.Fatal("expected respawned player position to be safe")
	}
}

func TestPlayerWithNoLivesCannotRespawn(t *testing.T) {
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
	game.playerSessions[playerID].Lives = 1
	game.state.Players[playerID].Lives = 1
	player := game.state.Players[playerID]
	game.state.Asteroids["asteroid-1"] = &entities.Asteroid{
		ID:   "asteroid-1",
		X:    player.X,
		Y:    player.Y,
		Size: 1,
	}

	game.handleShipAsteroidCollisions()

	if game.playerSessions[playerID].Lives != 0 {
		t.Fatalf("expected 0 lives after final death, got %d", game.playerSessions[playerID].Lives)
	}
	events := game.pendingEvents[playerID]
	if len(events) != 1 {
		t.Fatalf("expected 1 death event, got %d", len(events))
	}
	if events[0].Lives != 0 {
		t.Fatalf("expected game-over death event with 0 lives, got %d", events[0].Lives)
	}

	game.Step(constants.CollisionDespawnDelay)
	game.playerSessions[playerID].Step(constants.PlayerRespawnDelay)
	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeRespawn})
	if _, ok := game.state.Players[playerID]; ok {
		t.Fatal("expected respawn to be blocked with no lives")
	}
}
