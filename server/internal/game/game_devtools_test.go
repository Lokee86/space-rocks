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

func TestDebugFreezeWorldToggleCanBeDisabled(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugFreezeWorld})
	if !game.worldDevTools.IsWorldFrozen() {
		t.Fatal("expected first toggle to freeze world")
	}

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugFreezeWorld})
	if game.worldDevTools.IsWorldFrozen() {
		t.Fatal("expected second toggle to unfreeze world")
	}
}

func TestDebugFrozenWorldDoesNotMoveAsteroids(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	game.state.Asteroids["asteroid-1"] = &entities.Asteroid{
		ID:       "asteroid-1",
		X:        10,
		Y:        20,
		Velocity: physics.Vector2{X: 100, Y: 50},
		Size:     1,
	}

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugFreezeWorld})
	game.Step(1)

	asteroid := game.state.Asteroids["asteroid-1"]
	if asteroid.X != 10 || asteroid.Y != 20 {
		t.Fatalf("expected frozen asteroid to stay at (10, 20), got (%v, %v)", asteroid.X, asteroid.Y)
	}
}

func TestDebugFrozenWorldDoesNotMoveOrExpireBullets(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	game.state.Projectiles["bullet-1"] = entities.NewBullet(
		"bullet-1",
		playerID,
		physics.Vector2{X: 10, Y: 20},
		0,
		physics.Vector2{X: 100, Y: 50},
	)
	bullet := game.state.Projectiles["bullet-1"]
	startLife := bullet.Life

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugFreezeWorld})
	game.Step(startLife + 1)

	if bullet.X != 10 || bullet.Y != 20 {
		t.Fatalf("expected frozen bullet to stay at (10, 20), got (%v, %v)", bullet.X, bullet.Y)
	}
	if bullet.Life != startLife {
		t.Fatalf("expected frozen bullet life to stay %v, got %v", startLife, bullet.Life)
	}
	if _, ok := game.state.Projectiles["bullet-1"]; !ok {
		t.Fatal("expected frozen bullet not to expire")
	}
}

func TestDebugFrozenWorldDoesNotSpawnBullets(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugFreezeWorld})
	game.HandlePacket(playerID, ClientPacket{
		Type:  PacketTypeInput,
		Input: entities.InputState{Shoot: true},
	})
	game.Step(1.0 / float64(constants.ServerTickRate))

	if len(game.state.Projectiles) != 0 {
		t.Fatalf("expected frozen world not to spawn bullets, got %d", len(game.state.Projectiles))
	}
}

func TestDebugFrozenWorldDoesNotSpawnAsteroids(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	game.asteroidSpawnElapsed = constants.AsteroidSpawnInterval

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugFreezeWorld})
	game.Step(constants.AsteroidSpawnInterval)

	if len(game.state.Asteroids) != 0 {
		t.Fatalf("expected frozen world not to spawn asteroids, got %d", len(game.state.Asteroids))
	}
	if game.asteroidSpawnElapsed != constants.AsteroidSpawnInterval {
		t.Fatalf("expected frozen spawn timer not to advance, got %v", game.asteroidSpawnElapsed)
	}
}

func TestDebugFrozenWorldDoesNotRunShipAsteroidCollisions(t *testing.T) {
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

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugFreezeWorld})
	game.Step(1.0 / float64(constants.ServerTickRate))

	if player.IsPendingDespawn() {
		t.Fatal("expected frozen world not to mark colliding player for despawn")
	}
	if game.playerSessions[playerID].Lives != constants.PlayerStartingLives {
		t.Fatalf("expected frozen world to preserve %d lives, got %d", constants.PlayerStartingLives, game.playerSessions[playerID].Lives)
	}
	if len(game.pendingEvents[playerID]) != 0 {
		t.Fatalf("expected no ship death events while frozen, got %d", len(game.pendingEvents[playerID]))
	}
}

func TestDebugFrozenWorldDoesNotRunBulletAsteroidCollisionsOrScore(t *testing.T) {
	game := New()
	game.collisionShapes = physics.CollisionShapeCatalog{
		Bullet: physics.ImportedCollisionShape{
			Type:   "circle",
			Radius: 5,
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
	game.state.Projectiles["bullet-1"] = entities.NewBullet(
		"bullet-1",
		playerID,
		player.Position(),
		0,
		physics.Vector2{},
	)
	game.state.Asteroids["asteroid-1"] = &entities.Asteroid{
		ID:   "asteroid-1",
		X:    player.X,
		Y:    player.Y,
		Size: 1,
	}

	game.HandlePacket(playerID, ClientPacket{Type: PacketTypeToggleDebugFreezeWorld})
	game.Step(1.0 / float64(constants.ServerTickRate))

	if game.state.Projectiles["bullet-1"].IsPendingDespawn() {
		t.Fatal("expected frozen world not to mark colliding bullet for despawn")
	}
	if game.state.Asteroids["asteroid-1"].IsPendingDespawn() {
		t.Fatal("expected frozen world not to mark hit asteroid for despawn")
	}
	if player.Score != 0 {
		t.Fatalf("expected no score while frozen, got %d", player.Score)
	}
	if len(game.pendingEvents[playerID]) != 0 {
		t.Fatalf("expected no bullet impact events while frozen, got %d", len(game.pendingEvents[playerID]))
	}
	if len(game.state.Asteroids) != 1 {
		t.Fatalf("expected no asteroid fragments while frozen, got %d asteroids", len(game.state.Asteroids))
	}
}
