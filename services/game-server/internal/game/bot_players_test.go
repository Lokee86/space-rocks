package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/teams"
)

func TestAddBotUsesPlayerInputAndRemovalLifecycle(t *testing.T) {
	gameInstance := NewWithSeed(7)
	botID := gameInstance.AddBot()

	gameInstance.mu.Lock()
	ship := gameInstance.entities.Players[botID]
	ship.X = 200
	ship.Y = 200
	ship.Rotation = 0
	gameInstance.entities.Asteroids["asteroid-1"] = &runtime.Asteroid{
		ID:   "asteroid-1",
		X:    200,
		Y:    80,
		Size: 1,
	}
	gameInstance.mu.Unlock()

	gameInstance.Step(1.0 / 60.0)

	gameInstance.mu.Lock()
	input := gameInstance.entities.Players[botID].Input
	movedY := gameInstance.entities.Players[botID].Y
	bulletCount := len(gameInstance.entities.Projectiles)
	_, controllerExists := gameInstance.botControllers[botID]
	gameInstance.mu.Unlock()
	if !controllerExists {
		t.Fatal("expected bot controller to remain registered")
	}
	if !input.Forward || !input.PrimaryFire {
		t.Fatalf("expected normal player input from bot, got %+v", input)
	}
	if movedY >= 200 {
		t.Fatalf("expected bot ship to move toward asteroid, y=%f", movedY)
	}
	if bulletCount == 0 {
		t.Fatal("expected bot fire input to create a normal projectile")
	}

	gameInstance.RemovePlayer(botID)
	gameInstance.mu.Lock()
	_, playerExists := gameInstance.entities.Players[botID]
	_, controllerExists = gameInstance.botControllers[botID]
	gameInstance.mu.Unlock()
	if playerExists || controllerExists {
		t.Fatalf("expected bot player and controller removal, player=%v controller=%v", playerExists, controllerExists)
	}
}

func TestAddBotWithTeamPreservesMatchMembership(t *testing.T) {
	gameInstance := NewWithSeed(8)
	gameInstance.SetTeamStructure(teams.StructureCustom)
	botID := gameInstance.AddBotWithTeam(teams.Team3)

	if got := gameInstance.PlayerTeam(botID); got != teams.Team3 {
		t.Fatalf("bot team = %q, want %q", got, teams.Team3)
	}
	facts := gameInstance.PlayerMatchFacts()
	if len(facts) != 1 || facts[0].GamePlayerID != botID || facts[0].TeamID != teams.Team3 {
		t.Fatalf("bot facts = %+v, want Team3 membership", facts)
	}
}

func TestBotIgnoresAsteroidsOutsideLockedCameraView(t *testing.T) {
	gameInstance := NewWithSeed(9)
	botID := gameInstance.AddBot()

	gameInstance.mu.Lock()
	ship := gameInstance.entities.Players[botID]
	ship.X = 200
	ship.Y = 200
	ship.Rotation = 0
	gameInstance.cameraViews[botID].SetPosition(ship.Position())
	gameInstance.entities.Asteroids["offscreen"] = &runtime.Asteroid{
		ID:   "offscreen",
		X:    200,
		Y:    -800,
		Size: 1,
	}
	gameInstance.stepBots()
	input := ship.Input
	gameInstance.mu.Unlock()

	if !input.Forward {
		t.Fatalf("expected bot with no visible asteroid to continue forward, got %+v", input)
	}
	if input.PrimaryFire || input.Left || input.Right {
		t.Fatalf("expected off-screen asteroid to be absent from bot perception, got %+v", input)
	}
}

func TestBotRespawnsAfterNormalCooldown(t *testing.T) {
	gameInstance := NewWithSeed(11)
	botID := gameInstance.AddBot()

	gameInstance.mu.Lock()
	delete(gameInstance.entities.Players, botID)
	death := gameInstance.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: botID})
	gameInstance.mu.Unlock()
	if !death.Accepted {
		t.Fatalf("expected bot death transition, got %+v", death)
	}

	gameInstance.Step(gameInstance.lifeRuntime.Policy().RespawnDelay + 0.01)

	gameInstance.mu.Lock()
	ship := gameInstance.entities.Players[botID]
	gameInstance.mu.Unlock()
	if ship == nil {
		t.Fatal("expected bot to respawn through normal player lifecycle")
	}
}
