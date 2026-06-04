package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDebugInvincibleToggleCanBeDisabled(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInvincible,
	})
	if !scenario.playerInvincible(playerID) {
		t.Fatal("expected first toggle to make player invincible")
	}

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInvincible,
	})
	if scenario.playerInvincible(playerID) {
		t.Fatal("expected second toggle to make player vulnerable")
	}
}

func TestDebugInvincibleAllPlayersToggleAppliesToEveryPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInvincible,
	})
	if !scenario.playerInvincible(playerA) {
		t.Fatal("expected setup to make player A invincible")
	}
	if scenario.playerInvincible(playerB) {
		t.Fatal("expected setup to keep player B vulnerable")
	}

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeToggleDebugInvincible,
		TargetScope: "all_players",
	})

	if !scenario.playerInvincible(playerA) {
		t.Fatal("expected all-players invincible toggle to affect player A")
	}
	if !scenario.playerInvincible(playerB) {
		t.Fatal("expected all-players invincible toggle to affect player B")
	}

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeToggleDebugInvincible,
		TargetScope: "all_players",
	})

	if scenario.playerInvincible(playerA) {
		t.Fatal("expected second all-players invincible toggle to disable player A")
	}
	if scenario.playerInvincible(playerB) {
		t.Fatal("expected second all-players invincible toggle to disable player B")
	}
}

func TestStatePacketDebugStatusReflectsDebugToggles(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	initial := devtools.StatusFor(scenario.game, playerID)
	if initial.Invincible || initial.InfiniteLives || initial.WorldFrozen || initial.PlayerFrozen {
		t.Fatalf("expected initial debug status to be false, got %+v", initial)
	}

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInvincible,
	})
	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInfiniteLives,
	})
	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezePlayer,
	})

	status := devtools.StatusFor(scenario.game, playerID)
	if !status.Invincible {
		t.Fatal("expected debug status to report invincible")
	}
	if !status.InfiniteLives {
		t.Fatal("expected debug status to report infinite lives")
	}
	if !status.WorldFrozen {
		t.Fatal("expected debug status to report world frozen")
	}
	if !status.PlayerFrozen {
		t.Fatal("expected debug status to report player frozen")
	}
}

func TestDebugStatusReportsGranularWorldFreezeFlags(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "asteroids",
	})
	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "collisions",
	})

	status := devtools.StatusFor(scenario.game, playerID)
	if !status.AsteroidsFrozen {
		t.Fatal("expected debug status to report asteroids frozen")
	}
	if !status.CollisionsFrozen {
		t.Fatal("expected debug status to report collisions frozen")
	}
	if status.BulletsFrozen {
		t.Fatal("expected debug status to report bullets not frozen")
	}
	if status.SpawningFrozen {
		t.Fatal("expected debug status to report spawning not frozen")
	}
	if status.WorldFrozen {
		t.Fatal("expected debug status to report world not fully frozen")
	}
}

func TestDebugInvinciblePlayerDoesNotDieFromAsteroidCollision(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInvincible,
	})
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

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInfiniteLives,
	})
	if !scenario.playerInfiniteLives(playerID) {
		t.Fatal("expected first toggle to enable infinite lives")
	}

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInfiniteLives,
	})
	if scenario.playerInfiniteLives(playerID) {
		t.Fatal("expected second toggle to disable infinite lives")
	}
}

func TestDebugInfiniteLivesAllPlayersToggleAppliesToEveryPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInfiniteLives,
	})
	if !scenario.playerInfiniteLives(playerA) {
		t.Fatal("expected setup to enable infinite lives for player A")
	}
	if scenario.playerInfiniteLives(playerB) {
		t.Fatal("expected setup to keep player B finite")
	}

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeToggleDebugInfiniteLives,
		TargetScope: "all_players",
	})

	if !scenario.playerInfiniteLives(playerA) {
		t.Fatal("expected all-players infinite lives toggle to affect player A")
	}
	if !scenario.playerInfiniteLives(playerB) {
		t.Fatal("expected all-players infinite lives toggle to affect player B")
	}

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeToggleDebugInfiniteLives,
		TargetScope: "all_players",
	})

	if scenario.playerInfiniteLives(playerA) {
		t.Fatal("expected second all-players infinite lives toggle to disable player A")
	}
	if scenario.playerInfiniteLives(playerB) {
		t.Fatal("expected second all-players infinite lives toggle to disable player B")
	}
}

func TestDebugInfiniteLivesPlayerDiesWithoutLosingLife(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugInfiniteLives,
	})
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

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
	if !scenario.worldFrozen() {
		t.Fatal("expected first toggle to freeze world")
	}
	if !scenario.asteroidsFrozen() {
		t.Fatal("expected first toggle to freeze asteroids")
	}
	if !scenario.bulletsFrozen() {
		t.Fatal("expected first toggle to freeze bullets")
	}
	if !scenario.spawningFrozen() {
		t.Fatal("expected first toggle to freeze spawning")
	}
	if !scenario.collisionsFrozen() {
		t.Fatal("expected first toggle to freeze collisions")
	}

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
	if scenario.worldFrozen() {
		t.Fatal("expected second toggle to unfreeze world")
	}
	if scenario.asteroidsFrozen() {
		t.Fatal("expected second toggle to unfreeze asteroids")
	}
	if scenario.bulletsFrozen() {
		t.Fatal("expected second toggle to unfreeze bullets")
	}
	if scenario.spawningFrozen() {
		t.Fatal("expected second toggle to unfreeze spawning")
	}
	if scenario.collisionsFrozen() {
		t.Fatal("expected second toggle to unfreeze collisions")
	}
}

func TestDebugFreezePlayerAllPlayersToggleAppliesToEveryPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezePlayer,
	})
	if !devtools.StatusFor(scenario.game, playerA).PlayerFrozen {
		t.Fatal("expected setup to freeze player A")
	}
	if devtools.StatusFor(scenario.game, playerB).PlayerFrozen {
		t.Fatal("expected setup to keep player B unfrozen")
	}

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeToggleDebugFreezePlayer,
		TargetScope: "all_players",
	})

	if !devtools.StatusFor(scenario.game, playerA).PlayerFrozen {
		t.Fatal("expected all-players freeze player toggle to affect player A")
	}
	if !devtools.StatusFor(scenario.game, playerB).PlayerFrozen {
		t.Fatal("expected all-players freeze player toggle to affect player B")
	}

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeToggleDebugFreezePlayer,
		TargetScope: "all_players",
	})

	if devtools.StatusFor(scenario.game, playerA).PlayerFrozen {
		t.Fatal("expected second all-players freeze player toggle to unfreeze player A")
	}
	if devtools.StatusFor(scenario.game, playerB).PlayerFrozen {
		t.Fatal("expected second all-players freeze player toggle to unfreeze player B")
	}
}

func TestDebugFreezeWorldFromPartialFreezeEnablesAllFlags(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "asteroids",
	})
	if !scenario.asteroidsFrozen() {
		t.Fatal("expected asteroid-only freeze to enable asteroid freeze")
	}
	if scenario.worldFrozen() {
		t.Fatal("expected asteroid-only freeze not to mark world as fully frozen")
	}

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
	if !scenario.worldFrozen() {
		t.Fatal("expected aggregate freeze to fully freeze world from partial state")
	}
	if !scenario.asteroidsFrozen() {
		t.Fatal("expected aggregate freeze to keep asteroids frozen")
	}
	if !scenario.bulletsFrozen() {
		t.Fatal("expected aggregate freeze to freeze bullets")
	}
	if !scenario.spawningFrozen() {
		t.Fatal("expected aggregate freeze to freeze spawning")
	}
	if !scenario.collisionsFrozen() {
		t.Fatal("expected aggregate freeze to freeze collisions")
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

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
	scenario.step(1)

	asteroid := scenario.state(playerID).Asteroids["asteroid-1"]
	if asteroid.X != 10 || asteroid.Y != 20 {
		t.Fatalf("expected frozen asteroid to stay at (10, 20), got (%v, %v)", asteroid.X, asteroid.Y)
	}
}

func TestDebugFreezeAsteroidsOnlyStopsAsteroidMovement(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.placeMovingAsteroid(
		"asteroid-1",
		physics.Vector2{X: 10, Y: 20},
		physics.Vector2{X: 100, Y: 50},
		1,
	)
	scenario.placeBullet(
		"bullet-1",
		playerID,
		physics.Vector2{X: 200, Y: 300},
		physics.Vector2{X: 80, Y: 40},
	)

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "asteroids",
	})
	scenario.step(1)

	asteroid := scenario.state(playerID).Asteroids["asteroid-1"]
	if asteroid.X != 10 || asteroid.Y != 20 {
		t.Fatalf("expected asteroid-only freeze to keep asteroid at (10, 20), got (%v, %v)", asteroid.X, asteroid.Y)
	}

	bullet := scenario.state(playerID).Bullets["bullet-1"]
	if bullet.X == 200 && bullet.Y == 300 {
		t.Fatalf("expected bullet not to remain fully frozen at (200, 300), got (%v, %v)", bullet.X, bullet.Y)
	}

	if scenario.worldFrozen() {
		t.Fatal("expected asteroid-only freeze not to mark world fully frozen")
	}
	if !scenario.asteroidsFrozen() {
		t.Fatal("expected asteroid-only freeze to set asteroids frozen")
	}
	if scenario.bulletsFrozen() {
		t.Fatal("expected asteroid-only freeze not to freeze bullets")
	}
	if scenario.spawningFrozen() {
		t.Fatal("expected asteroid-only freeze not to freeze spawning")
	}
	if scenario.collisionsFrozen() {
		t.Fatal("expected asteroid-only freeze not to freeze collisions")
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

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
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

func TestDebugFreezeBulletsOnlyStopsBulletMovementAndExpiry(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.placeBullet(
		"bullet-1",
		playerID,
		physics.Vector2{X: 10, Y: 20},
		physics.Vector2{X: 100, Y: 50},
	)
	startLife := scenario.bulletLife("bullet-1")

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "bullets",
	})
	scenario.step(startLife + 1)

	bullet, ok := scenario.state(playerID).Bullets["bullet-1"]
	if !ok {
		t.Fatal("expected bullets-only freeze to keep bullet from expiring")
	}
	if bullet.X != 10 || bullet.Y != 20 {
		t.Fatalf("expected bullets-only freeze to keep bullet at (10, 20), got (%v, %v)", bullet.X, bullet.Y)
	}
	if life := scenario.bulletLife("bullet-1"); life != startLife {
		t.Fatalf("expected bullets-only freeze to keep bullet life at %v, got %v", startLife, life)
	}
	if !scenario.bulletsFrozen() {
		t.Fatal("expected bullets-only freeze to set bullets frozen")
	}
	if scenario.worldFrozen() {
		t.Fatal("expected bullets-only freeze not to mark world fully frozen")
	}
	if scenario.asteroidsFrozen() {
		t.Fatal("expected bullets-only freeze not to freeze asteroids")
	}
	if scenario.spawningFrozen() {
		t.Fatal("expected bullets-only freeze not to freeze spawning")
	}
	if scenario.collisionsFrozen() {
		t.Fatal("expected bullets-only freeze not to freeze collisions")
	}
}

func TestDebugFrozenWorldDoesNotSpawnBullets(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
	scenario.send(playerID, servergame.ClientPacket{
		Type:  servergame.PacketTypeInput,
		Input: runtime.InputState{Shoot: true},
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

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
	scenario.step(constants.AsteroidSpawnInterval)

	if asteroids := scenario.state(playerID).Asteroids; len(asteroids) != 0 {
		t.Fatalf("expected frozen world not to spawn asteroids, got %d", len(asteroids))
	}
	if elapsed := scenario.asteroidSpawnElapsed(); elapsed != constants.AsteroidSpawnInterval {
		t.Fatalf("expected frozen spawn timer not to advance, got %v", elapsed)
	}
}

func TestDebugFreezeSpawningOnlyStopsAsteroidSpawning(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	scenario.setAsteroidSpawnElapsed(constants.AsteroidSpawnInterval)

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "spawning",
	})
	scenario.step(constants.AsteroidSpawnInterval)

	if asteroids := scenario.state(playerID).Asteroids; len(asteroids) != 0 {
		t.Fatalf("expected spawning-only freeze not to spawn asteroids, got %d", len(asteroids))
	}
	if elapsed := scenario.asteroidSpawnElapsed(); elapsed != constants.AsteroidSpawnInterval {
		t.Fatalf("expected spawning-only freeze not to advance/reset spawn timer, got %v", elapsed)
	}
	if scenario.worldFrozen() {
		t.Fatal("expected spawning-only freeze not to mark world fully frozen")
	}
	if !scenario.spawningFrozen() {
		t.Fatal("expected spawning-only freeze to set spawning frozen")
	}
	if scenario.asteroidsFrozen() {
		t.Fatal("expected spawning-only freeze not to freeze asteroids")
	}
	if scenario.bulletsFrozen() {
		t.Fatal("expected spawning-only freeze not to freeze bullets")
	}
	if scenario.collisionsFrozen() {
		t.Fatal("expected spawning-only freeze not to freeze collisions")
	}
}

func TestDebugFreezeSpawnsAliasFreezesSpawning(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "spawns",
	})

	if !scenario.spawningFrozen() {
		t.Fatal("expected spawns alias to freeze spawning")
	}
	if scenario.worldFrozen() {
		t.Fatal("expected spawns alias not to mark world fully frozen")
	}
	if scenario.asteroidsFrozen() {
		t.Fatal("expected spawns alias not to freeze asteroids")
	}
	if scenario.bulletsFrozen() {
		t.Fatal("expected spawns alias not to freeze bullets")
	}
	if scenario.collisionsFrozen() {
		t.Fatal("expected spawns alias not to freeze collisions")
	}
}

func TestDebugFreezeUnknownTargetDoesNotChangeFreezeFlags(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "bogus",
	})

	if scenario.worldFrozen() {
		t.Fatal("expected unknown freeze target not to mark world fully frozen")
	}
	if scenario.asteroidsFrozen() {
		t.Fatal("expected unknown freeze target not to freeze asteroids")
	}
	if scenario.bulletsFrozen() {
		t.Fatal("expected unknown freeze target not to freeze bullets")
	}
	if scenario.spawningFrozen() {
		t.Fatal("expected unknown freeze target not to freeze spawning")
	}
	if scenario.collisionsFrozen() {
		t.Fatal("expected unknown freeze target not to freeze collisions")
	}
}

func TestDebugFrozenWorldDoesNotRunShipAsteroidCollisions(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
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

func TestDebugFreezeCollisionsOnlyStopsCollisionConsequences(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "collisions",
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected collisions-only freeze not to mark colliding player for despawn")
	}
	if packet := scenario.state(playerID); packet.Lives != constants.PlayerStartingLives {
		t.Fatalf("expected collisions-only freeze to preserve %d lives, got %d", constants.PlayerStartingLives, packet.Lives)
	}
	if events := scenario.pendingEventCount(playerID); events != 0 {
		t.Fatalf("expected no ship death events while collisions are frozen, got %d", events)
	}
	if scenario.worldFrozen() {
		t.Fatal("expected collisions-only freeze not to mark world fully frozen")
	}
	if !scenario.collisionsFrozen() {
		t.Fatal("expected collisions-only freeze to set collisions frozen")
	}
	if scenario.asteroidsFrozen() {
		t.Fatal("expected collisions-only freeze not to freeze asteroids")
	}
	if scenario.bulletsFrozen() {
		t.Fatal("expected collisions-only freeze not to freeze bullets")
	}
	if scenario.spawningFrozen() {
		t.Fatal("expected collisions-only freeze not to freeze spawning")
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

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeToggleDebugFreezeWorld,
	})
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

func TestDebugKillPlayerMarksDespawnQueuesDeathAndReducesLives(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type: devtools.PacketTypeDebugKillPlayer,
	})

	if !scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected debug kill to mark player pending despawn")
	}
	if events := scenario.pendingEventCount(playerID); events != 1 {
		t.Fatalf("expected one queued ship death event, got %d", events)
	}
	packet := scenario.state(playerID)
	if len(packet.Events) != 1 {
		t.Fatalf("expected one ship death event in packet, got %d", len(packet.Events))
	}
	if packet.Events[0].Type != servergame.PacketTypeShipDeath {
		t.Fatalf("expected ship death event type %q, got %q", servergame.PacketTypeShipDeath, packet.Events[0].Type)
	}
	expectedLives := constants.PlayerStartingLives - 1
	if packet.Lives != expectedLives {
		t.Fatalf("expected debug kill to reduce lives to %d, got %d", expectedLives, packet.Lives)
	}
}

func TestDebugKillPlayerCanKillAnotherActivePlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:           devtools.PacketTypeDebugKillPlayer,
		TargetPlayerID: playerB,
	})

	if scenario.playerPendingDespawn(playerA) {
		t.Fatal("expected source player to remain active")
	}
	if !scenario.playerPendingDespawn(playerB) {
		t.Fatal("expected target player to be marked pending despawn")
	}
	packetA := scenario.state(playerA)
	if len(packetA.Events) != 1 {
		t.Fatalf("expected one ship death event in source view, got %d", len(packetA.Events))
	}
	if packetA.Events[0].Type != servergame.PacketTypeShipDeath {
		t.Fatalf("expected ship death event type %q, got %q", servergame.PacketTypeShipDeath, packetA.Events[0].Type)
	}
	if packetA.Events[0].PlayerID != playerB {
		t.Fatalf("expected ship death event player id %q, got %q", playerB, packetA.Events[0].PlayerID)
	}
	packetB := scenario.state(playerB)
	expectedLives := constants.PlayerStartingLives - 1
	if packetB.Lives != expectedLives {
		t.Fatalf("expected target debug kill to reduce lives to %d, got %d", expectedLives, packetB.Lives)
	}
}

func TestDebugKillPlayerAllPlayersAppliesToEveryPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeDebugKillPlayer,
		TargetScope: "all_players",
	})

	if !scenario.playerPendingDespawn(playerA) {
		t.Fatal("expected all-players debug kill to mark player A pending despawn")
	}
	if !scenario.playerPendingDespawn(playerB) {
		t.Fatal("expected all-players debug kill to mark player B pending despawn")
	}
}

func TestDebugSetScoreAllPlayersAppliesToEveryPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeDebugSetScore,
		TargetScope: "all_players",
		Score:       44,
	})

	if score := scenario.playerState(playerA, playerA).Score; score != 44 {
		t.Fatalf("expected player A score 44, got %d", score)
	}
	if score := scenario.playerState(playerB, playerB).Score; score != 44 {
		t.Fatalf("expected player B score 44, got %d", score)
	}
}

func TestSetPlayerScoreExportsSessionOwnedScoreInStatePacket(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	change := scenario.game.SetPlayerScore(playerID, 37)
	if !change.Found {
		t.Fatalf("expected SetPlayerScore to find player %q", playerID)
	}

	player := scenario.playerState(playerID, playerID)
	if player.Score != 37 {
		t.Fatalf("expected state packet score 37, got %d", player.Score)
	}
}

func TestSetPlayerLivesExportsSessionOwnedLivesInStatePacket(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	change := scenario.game.SetPlayerLives(playerID, 6)
	if !change.Found {
		t.Fatalf("expected SetPlayerLives to find player %q", playerID)
	}

	packet := scenario.state(playerID)
	player, ok := packet.Players[playerID]
	if !ok {
		t.Fatalf("expected state packet for %q to include player %q", playerID, playerID)
	}
	if player.Lives != 6 {
		t.Fatalf("expected state packet player lives 6, got %d", player.Lives)
	}
	if packet.Lives != 6 {
		t.Fatalf("expected top-level packet lives 6, got %d", packet.Lives)
	}
}

func TestDebugAddScoreAllPlayersAppliesToEveryPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeDebugSetScore,
		TargetScope: "all_players",
		Score:       10,
	})
	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeDebugAddScore,
		TargetScope: "all_players",
		Amount:      6,
	})

	if score := scenario.playerState(playerA, playerA).Score; score != 16 {
		t.Fatalf("expected player A score 16, got %d", score)
	}
	if score := scenario.playerState(playerB, playerB).Score; score != 16 {
		t.Fatalf("expected player B score 16, got %d", score)
	}
}

func TestDebugSetLivesAllPlayersAppliesToEveryPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeDebugSetLives,
		TargetScope: "all_players",
		Lives:       7,
	})

	if lives := scenario.state(playerA).Lives; lives != 7 {
		t.Fatalf("expected player A packet lives 7, got %d", lives)
	}
	if lives := scenario.state(playerB).Lives; lives != 7 {
		t.Fatalf("expected player B packet lives 7, got %d", lives)
	}
}

func TestDebugAddLivesAllPlayersAppliesToEveryPlayer(t *testing.T) {
	scenario := newScenario(t)
	playerA := scenario.addPlayer()
	playerB := scenario.addPlayer()

	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeDebugSetLives,
		TargetScope: "all_players",
		Lives:       3,
	})
	devtools.HandleCommand(scenario.game, playerA, devtools.DebugCommand{
		Type:        devtools.PacketTypeDebugAddLives,
		TargetScope: "all_players",
		Amount:      2,
	})

	if lives := scenario.state(playerA).Lives; lives != 5 {
		t.Fatalf("expected player A packet lives 5, got %d", lives)
	}
	if lives := scenario.state(playerB).Lives; lives != 5 {
		t.Fatalf("expected player B packet lives 5, got %d", lives)
	}
}

func TestDebugRespawnPlayerAllPlayersRespawnsEligiblePlayersAndIgnoresActivePlayers(t *testing.T) {
	scenario := newScenario(t)
	activePlayerID := scenario.addPlayer()
	respawnEligiblePlayerID := scenario.addPlayer()
	activePlayerBefore := scenario.playerState(activePlayerID, activePlayerID)
	respawnPosition := physics.Vector2{X: 321, Y: 654}

	scenario.removePlayerEntity(respawnEligiblePlayerID)
	scenario.setSessionSpawnPosition(respawnEligiblePlayerID, respawnPosition)

	devtools.HandleCommand(scenario.game, activePlayerID, devtools.DebugCommand{
		Type:        devtools.PacketTypeDebugRespawnPlayer,
		TargetScope: "all_players",
	})

	if !scenario.playerEntityExists(respawnEligiblePlayerID) {
		t.Fatal("expected all-players debug respawn to recreate the eligible player entity")
	}
	respawned := scenario.playerState(activePlayerID, respawnEligiblePlayerID)
	if respawned.X != respawnPosition.X || respawned.Y != respawnPosition.Y {
		t.Fatalf(
			"expected eligible player to respawn at (%v, %v), got (%v, %v)",
			respawnPosition.X,
			respawnPosition.Y,
			respawned.X,
			respawned.Y,
		)
	}

	activePlayerAfter := scenario.playerState(activePlayerID, activePlayerID)
	if activePlayerAfter.X != activePlayerBefore.X || activePlayerAfter.Y != activePlayerBefore.Y {
		t.Fatalf(
			"expected active player to be ignored by respawn guard and stay at (%v, %v), got (%v, %v)",
			activePlayerBefore.X,
			activePlayerBefore.Y,
			activePlayerAfter.X,
			activePlayerAfter.Y,
		)
	}
}
