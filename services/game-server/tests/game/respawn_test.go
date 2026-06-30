package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestPlayerDeathReducesLivesAndAllowsRespawnAfterDelay(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	spawnPosition := physics.Vector2{X: player.X, Y: player.Y}
	scenario.placeAsteroid("asteroid-1", spawnPosition, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	snapshot := scenario.presentationSnapshot(playerID)
	if snapshot.Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected gameplay snapshot overlay/session lives %d after death, got %d", constants.PlayerStartingLives-1, snapshot.Lives)
	}
	if countPendingEventsOfType(snapshot.PendingEvents, servergame.PacketTypeShipDeath) != 1 {
		t.Fatalf("expected 1 death event, got %d", countPendingEventsOfType(snapshot.PendingEvents, servergame.PacketTypeShipDeath))
	}
	var deathEvent *servergame.EventState
	for i := range snapshot.PendingEvents {
		if snapshot.PendingEvents[i].Event.Type == servergame.PacketTypeShipDeath {
			deathEvent = &snapshot.PendingEvents[i].Event
			break
		}
	}
	if deathEvent == nil {
		t.Fatal("expected ship_death event in gameplay snapshot")
	}
	if deathEvent.Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected death event lives %d, got %d", constants.PlayerStartingLives-1, deathEvent.Lives)
	}
	if deathEvent.RespawnDelay != constants.PlayerRespawnDelay {
		t.Fatalf("expected respawn delay %v, got %v", constants.PlayerRespawnDelay, deathEvent.RespawnDelay)
	}
	foundDamageApplied := false
	for _, pending := range snapshot.PendingEvents {
		if pending.Event.Type == "damage_applied" {
			foundDamageApplied = true
			break
		}
	}
	if !foundDamageApplied {
		t.Fatal("expected damage_applied event for asteroid collision death")
	}

	scenario.step(constants.CollisionDespawnDelay)
	if scenario.playerExists(playerID, playerID) {
		t.Fatal("expected dead player entity to be removed before respawn")
	}
	if snapshot := scenario.presentationSnapshot(playerID); snapshot.Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected gameplay snapshot overlay/session lives %d after player removal, got %d", constants.PlayerStartingLives-1, snapshot.Lives)
	}

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})
	if scenario.playerExists(playerID, playerID) {
		t.Fatal("expected respawn to be blocked before delay finishes")
	}

	scenario.advanceRespawnTimer(playerID, constants.PlayerRespawnDelay)
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	respawned := scenario.playerState(playerID, playerID)
	respawnedSession := scenario.playerSessionState(playerID, playerID)
	if respawnedSession.Lives != constants.PlayerStartingLives-1 {
		t.Fatalf("expected respawned player to keep %d lives, got %d", constants.PlayerStartingLives-1, respawnedSession.Lives)
	}
	if respawned.X == spawnPosition.X && respawned.Y == spawnPosition.Y {
		t.Fatal("expected respawn to avoid asteroid on original spawn point")
	}
}

func TestAddedLivesPersistThroughDeathAndRespawn(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	change := scenario.game.AddPlayerLives(playerID, 2)
	if !change.Found {
		t.Fatalf("expected AddPlayerLives to find player %q", playerID)
	}

	initial := scenario.playerState(playerID, playerID)
	spawnPosition := physics.Vector2{X: initial.X, Y: initial.Y}
	scenario.placeAsteroid("asteroid-1", spawnPosition, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))
	scenario.step(constants.CollisionDespawnDelay)
	scenario.advanceRespawnTimer(playerID, constants.PlayerRespawnDelay)
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	respawnedSession := scenario.playerSessionState(playerID, playerID)
	expectedLives := constants.PlayerStartingLives + 1
	if respawnedSession.Lives != expectedLives {
		t.Fatalf("expected respawned player to keep %d lives, got %d", expectedLives, respawnedSession.Lives)
	}
	if snapshot := scenario.presentationSnapshot(playerID); snapshot.Lives != expectedLives {
		t.Fatalf("expected gameplay snapshot overlay/session lives %d after respawn, got %d", expectedLives, snapshot.Lives)
	}
}

func TestPlayerWithNoLivesCannotRespawn(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.setPlayerLives(playerID, 1)
	player := scenario.playerState(playerID, playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	snapshot := scenario.presentationSnapshot(playerID)
	if snapshot.Lives != 0 {
		t.Fatalf("expected 0 lives after final death, got %d", snapshot.Lives)
	}
	if countPendingEventsOfType(snapshot.PendingEvents, servergame.PacketTypeShipDeath) != 1 {
		t.Fatalf("expected 1 death event, got %d", countPendingEventsOfType(snapshot.PendingEvents, servergame.PacketTypeShipDeath))
	}
	var deathEvent *servergame.EventState
	for i := range snapshot.PendingEvents {
		if snapshot.PendingEvents[i].Event.Type == servergame.PacketTypeShipDeath {
			deathEvent = &snapshot.PendingEvents[i].Event
			break
		}
	}
	if deathEvent == nil {
		t.Fatal("expected ship_death event in gameplay snapshot")
	}
	if deathEvent.Lives != 0 {
		t.Fatalf("expected game-over death event with 0 lives, got %d", deathEvent.Lives)
	}
	foundDamageApplied := false
	for _, pending := range snapshot.PendingEvents {
		if pending.Event.Type == "damage_applied" {
			foundDamageApplied = true
			break
		}
	}
	if !foundDamageApplied {
		t.Fatal("expected damage_applied event for asteroid collision death")
	}

	scenario.step(constants.CollisionDespawnDelay)
	scenario.advanceRespawnTimer(playerID, constants.PlayerRespawnDelay)
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})
	if !scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected respawn to remain blocked with no lives")
	}
	if snapshot := scenario.presentationSnapshot(playerID); snapshot.Lives != 0 {
		t.Fatalf("expected no-lives respawn attempt to keep snapshot lives 0, got %d", snapshot.Lives)
	}
	if lives := scenario.playerSessionState(playerID, playerID).Lives; lives != 0 {
		t.Fatalf("expected no-lives respawn attempt to keep session lives 0, got %d", lives)
	}
}

func TestRespawnSafetyUsesRespawnBuffer(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.removePlayerEntity(playerID)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{}, 1)

	insideBuffer := physics.Vector2{X: constants.PlayerRespawnBuffer + 21, Y: 0}
	scenario.setSessionSpawnPosition(playerID, insideBuffer)
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	respawned := scenario.playerState(playerID, playerID)
	if respawned.X == insideBuffer.X && respawned.Y == insideBuffer.Y {
		t.Fatal("expected respawn position inside asteroid buffer to be avoided")
	}

	outsideScenario := newScenario(t)
	outsideScenario.useCircleCollisionShapes()
	outsidePlayerID := outsideScenario.addPlayer()
	outsideScenario.removePlayerEntity(outsidePlayerID)
	outsideScenario.placeAsteroid("asteroid-1", physics.Vector2{}, 1)

	outsideBuffer := physics.Vector2{X: constants.PlayerRespawnBuffer + 128, Y: 0}
	outsideScenario.setSessionSpawnPosition(outsidePlayerID, outsideBuffer)
	outsideScenario.send(outsidePlayerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	outsideRespawned := outsideScenario.playerState(outsidePlayerID, outsidePlayerID)
	if outsideRespawned.X != outsideBuffer.X || outsideRespawned.Y != outsideBuffer.Y {
		t.Fatalf("expected respawn position outside asteroid buffer to be used, got (%v, %v)", outsideRespawned.X, outsideRespawned.Y)
	}
}

func TestRespawnSafetySeesAsteroidAcrossWrapBoundary(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.removePlayerEntity(playerID)
	spawnPosition := physics.Vector2{X: 5, Y: 100}
	scenario.setSessionSpawnPosition(playerID, spawnPosition)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: constants.WorldWidth - 5, Y: 100}, 1)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	respawned := scenario.playerState(playerID, playerID)
	if respawned.X == spawnPosition.X && respawned.Y == spawnPosition.Y {
		t.Fatal("expected respawn position near cross-edge asteroid to be avoided")
	}
}

func TestRespawnSafetySeesPlayerAcrossWrapBoundary(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	respawningPlayerID := scenario.addPlayer()
	otherPlayerID := scenario.addPlayer()
	spawnPosition := physics.Vector2{X: 5, Y: 100}
	scenario.setSessionSpawnPosition(respawningPlayerID, spawnPosition)
	scenario.setPlayerPosition(otherPlayerID, physics.Vector2{X: constants.WorldWidth - 5, Y: 100})
	scenario.removePlayerEntity(respawningPlayerID)

	scenario.send(respawningPlayerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	respawned := scenario.playerState(respawningPlayerID, respawningPlayerID)
	if respawned.X == spawnPosition.X && respawned.Y == spawnPosition.Y {
		t.Fatal("expected respawn position near cross-edge player to be avoided")
	}
}

func TestRespawnSearchFindsSafePointAfterWrappedUnsafeCandidates(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.removePlayerEntity(playerID)
	spawnPosition := physics.Vector2{X: 5, Y: 100}
	firstSearchCandidate := physics.Vector2{
		X: spawnPosition.X - constants.PlayerRespawnBuffer,
		Y: spawnPosition.Y - constants.PlayerRespawnBuffer,
	}
	scenario.setSessionSpawnPosition(playerID, spawnPosition)
	scenario.placeAsteroid("asteroid-origin", physics.Vector2{X: constants.WorldWidth - 5, Y: 100}, 1)
	scenario.placeAsteroid("asteroid-first-candidate", physics.Vector2{
		X: constants.WorldWidth - constants.PlayerRespawnBuffer + spawnPosition.X,
		Y: constants.WorldHeight - constants.PlayerRespawnBuffer + spawnPosition.Y,
	}, 1)

	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	respawned := scenario.playerState(playerID, playerID)
	if respawned.X == spawnPosition.X && respawned.Y == spawnPosition.Y {
		t.Fatal("expected wrapped-unsafe origin to be avoided")
	}
	if respawned.X == firstSearchCandidate.X && respawned.Y == firstSearchCandidate.Y {
		t.Fatal("expected wrapped-unsafe first search candidate to be avoided")
	}
}

func TestRespawnSafetyAvoidsExistingPlayers(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	respawningPlayerID := scenario.addPlayer()
	otherPlayerID := scenario.addPlayer()
	scenario.setPlayerPosition(otherPlayerID, physics.Vector2{})
	scenario.removePlayerEntity(respawningPlayerID)

	insideBuffer := physics.Vector2{X: constants.PlayerRespawnBuffer + 39, Y: 0}
	scenario.setSessionSpawnPosition(respawningPlayerID, insideBuffer)
	scenario.send(respawningPlayerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	respawned := scenario.playerState(respawningPlayerID, respawningPlayerID)
	if respawned.X == insideBuffer.X && respawned.Y == insideBuffer.Y {
		t.Fatal("expected respawn position inside player buffer to be avoided")
	}

	outsideScenario := newScenario(t)
	outsideScenario.useCircleCollisionShapes()
	outsideRespawningPlayerID := outsideScenario.addPlayer()
	outsideOtherPlayerID := outsideScenario.addPlayer()
	outsideScenario.setPlayerPosition(outsideOtherPlayerID, physics.Vector2{})
	outsideScenario.removePlayerEntity(outsideRespawningPlayerID)

	outsideBuffer := physics.Vector2{X: constants.PlayerRespawnBuffer + 41, Y: 0}
	outsideScenario.setSessionSpawnPosition(outsideRespawningPlayerID, outsideBuffer)
	outsideScenario.send(outsideRespawningPlayerID, servergame.ClientPacket{Type: servergame.PacketTypeRespawn})

	outsideRespawned := outsideScenario.playerState(outsideRespawningPlayerID, outsideRespawningPlayerID)
	if outsideRespawned.X != outsideBuffer.X || outsideRespawned.Y != outsideBuffer.Y {
		t.Fatalf("expected respawn position outside player buffer to be used, got (%v, %v)", outsideRespawned.X, outsideRespawned.Y)
	}
}

func TestInitialSpawnAvoidsExistingPlayers(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()

	firstID := scenario.addPlayer()
	secondID := scenario.addPlayer()
	first := scenario.playerState(firstID, firstID)
	second := scenario.playerState(secondID, secondID)
	secondSessionSpawn := scenario.sessionSpawnPosition(secondID)

	if firstID == secondID {
		t.Fatal("expected unique player IDs")
	}
	if first.X == second.X && first.Y == second.Y {
		t.Fatal("expected initial spawn to avoid existing player")
	}
	if secondSessionSpawn.X != second.X || secondSessionSpawn.Y != second.Y {
		t.Fatal("expected session spawn position to match safe initial spawn")
	}
}
