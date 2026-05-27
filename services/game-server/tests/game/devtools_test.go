package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDebugInvincibleToggleCanBeDisabled(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugInvincible})
	if !scenario.playerInvincible(playerID) {
		t.Fatal("expected first toggle to make player invincible")
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugInvincible})
	if scenario.playerInvincible(playerID) {
		t.Fatal("expected second toggle to make player vulnerable")
	}
}

func TestDebugInvinciblePlayerDoesNotDieFromAsteroidCollision(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugInvincible})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected invincible player not to be marked for despawn")
	}
	if lives := scenario.state(playerID).Lives; lives != constants.PlayerStartingLives {
		t.Fatalf("expected invincible player to keep %d lives, got %d", constants.PlayerStartingLives, lives)
	}
	if events := scenario.pendingEventCount(playerID); events != 0 {
		t.Fatalf("expected no death events for invincible player, got %d", events)
	}
}

func TestDebugInfiniteLivesToggleCanBeDisabled(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugInfiniteLives})
	if !scenario.playerInfiniteLives(playerID) {
		t.Fatal("expected first toggle to enable infinite lives")
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugInfiniteLives})
	if scenario.playerInfiniteLives(playerID) {
		t.Fatal("expected second toggle to disable infinite lives")
	}
}

func TestDebugInfiniteLivesPlayerDiesWithoutLosingLife(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugInfiniteLives})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if !scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected infinite-lives player to still die and despawn")
	}
	packet := scenario.state(playerID)
	if packet.Lives != constants.PlayerStartingLives {
		t.Fatalf("expected infinite-lives player to keep %d lives, got %d", constants.PlayerStartingLives, packet.Lives)
	}
	if len(packet.Events) != 1 {
		t.Fatalf("expected death event for infinite-lives player, got %d", len(packet.Events))
	}
	if packet.Events[0].Lives != constants.PlayerStartingLives {
		t.Fatalf("expected death event to keep %d lives, got %d", constants.PlayerStartingLives, packet.Events[0].Lives)
	}

	scenario.step(constants.CollisionDespawnDelay)
	scenario.advanceRespawnTimer(playerID, constants.PlayerRespawnDelay)
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	if !scenario.playerInfiniteLives(playerID) {
		t.Fatal("expected infinite lives flag to persist after respawn")
	}
}

func TestDebugFreezeWorldToggleCanBeDisabled(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugFreezeWorld})
	if !scenario.worldFrozen() {
		t.Fatal("expected first toggle to freeze world")
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugFreezeWorld})
	if scenario.worldFrozen() {
		t.Fatal("expected second toggle to unfreeze world")
	}
}

func TestDebugFrozenWorldDoesNotMoveAsteroids(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.placeMovingAsteroid(
		"asteroid-1",
		physics.Vector2{X: 10, Y: 20},
		physics.Vector2{X: 100, Y: 50},
		1,
	)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugFreezeWorld})
	scenario.step(1)

	asteroid := scenario.state(playerID).Asteroids["asteroid-1"]
	if asteroid.X != 10 || asteroid.Y != 20 {
		t.Fatalf("expected frozen asteroid to stay at (10, 20), got (%v, %v)", asteroid.X, asteroid.Y)
	}
}

func TestDebugFrozenWorldDoesNotMoveOrExpireBullets(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.placeBullet(
		"bullet-1",
		playerID,
		physics.Vector2{X: 10, Y: 20},
		physics.Vector2{X: 100, Y: 50},
	)
	startLife := scenario.bulletLife("bullet-1")

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugFreezeWorld})
	scenario.step(startLife + 1)

	bullet, ok := scenario.state(playerID).Bullets["bullet-1"]
	if !ok {
		t.Fatal("expected frozen bullet not to expire")
	}
	if bullet.X != 10 || bullet.Y != 20 {
		t.Fatalf("expected frozen bullet to stay at (10, 20), got (%v, %v)", bullet.X, bullet.Y)
	}
	if life := scenario.bulletLife("bullet-1"); life != startLife {
		t.Fatalf("expected frozen bullet life to stay %v, got %v", startLife, life)
	}
}

func TestDebugFrozenWorldDoesNotSpawnBullets(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugFreezeWorld})
	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: entities.InputState{Shoot: true},
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if bullets := scenario.state(playerID).Bullets; len(bullets) != 0 {
		t.Fatalf("expected frozen world not to spawn bullets, got %d", len(bullets))
	}
}

func TestDebugFrozenWorldDoesNotSpawnAsteroids(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.setAsteroidSpawnElapsed(constants.AsteroidSpawnInterval)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugFreezeWorld})
	scenario.step(constants.AsteroidSpawnInterval)

	if asteroids := scenario.state(playerID).Asteroids; len(asteroids) != 0 {
		t.Fatalf("expected frozen world not to spawn asteroids, got %d", len(asteroids))
	}
	if elapsed := scenario.asteroidSpawnElapsed(); elapsed != constants.AsteroidSpawnInterval {
		t.Fatalf("expected frozen spawn timer not to advance, got %v", elapsed)
	}
}

func TestDebugFrozenWorldDoesNotRunShipAsteroidCollisions(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugFreezeWorld})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected frozen world not to mark colliding player for despawn")
	}
	if packet := scenario.state(playerID); packet.Lives != constants.PlayerStartingLives {
		t.Fatalf("expected frozen world to preserve %d lives, got %d", constants.PlayerStartingLives, packet.Lives)
	}
	if events := scenario.pendingEventCount(playerID); events != 0 {
		t.Fatalf("expected no ship death events while frozen, got %d", events)
	}
}

func TestDebugFrozenWorldDoesNotRunBulletAsteroidCollisionsOrScore(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	position := physics.Vector2{X: player.X, Y: player.Y}
	scenario.placeBullet("bullet-1", playerID, position, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", position, 1)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeToggleDebugFreezeWorld})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if scenario.bulletPendingDespawn("bullet-1") {
		t.Fatal("expected frozen world not to mark colliding bullet for despawn")
	}
	if scenario.asteroidPendingDespawn("asteroid-1") {
		t.Fatal("expected frozen world not to mark hit asteroid for despawn")
	}
	if player := scenario.playerState(playerID, playerID); player.Score != 0 {
		t.Fatalf("expected no score while frozen, got %d", player.Score)
	}
	if events := scenario.pendingEventCount(playerID); events != 0 {
		t.Fatalf("expected no bullet impact events while frozen, got %d", events)
	}
	if asteroids := scenario.state(playerID).Asteroids; len(asteroids) != 1 {
		t.Fatalf("expected no asteroid fragments while frozen, got %d asteroids", len(asteroids))
	}
}
