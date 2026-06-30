package gametests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	servergame "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestBulletAsteroidCollisionsDelayHitDespawns(t *testing.T) {
	scenario := newScenario(t)
	scenario.useBulletCapsuleAsteroidPolygonCollisions()
	playerID := scenario.addPlayer()
	impactPosition := physics.Vector2{X: 100, Y: 100}
	scenario.placeBullet("bullet-1", playerID, impactPosition, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", impactPosition, 1)
	scenario.placeAsteroid("asteroid-2", physics.Vector2{X: 220, Y: 100}, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	snapshot := scenario.presentationSnapshot(playerID)
	if _, ok := snapshot.Bullets["bullet-1"]; !ok {
		t.Fatal("expected hit bullet to remain during despawn delay")
	}
	if !scenario.bulletPendingDespawn("bullet-1") {
		t.Fatal("expected hit bullet to be marked for delayed despawn")
	}
	if _, ok := snapshot.Asteroids["asteroid-1"]; !ok {
		t.Fatal("expected hit asteroid to remain during despawn delay")
	}
	if !scenario.asteroidPendingDespawn("asteroid-1") {
		t.Fatal("expected hit asteroid to be marked for delayed despawn")
	}
	if _, ok := snapshot.Asteroids["asteroid-2"]; !ok {
		t.Fatal("expected untouched asteroid to remain")
	}
	if len(snapshot.PendingEvents) != 2 {
		t.Fatalf("expected 2 events in gameplay snapshot event batch, got %d", len(snapshot.PendingEvents))
	}
	foundBulletBlast := false
	foundDamageApplied := false
	for _, pending := range snapshot.PendingEvents {
		event := pending.Event
		switch event.Type {
		case servergame.PacketTypeBulletBlast:
			foundBulletBlast = true
		case "damage_applied":
			foundDamageApplied = true
		}
	}
	if !foundBulletBlast {
		t.Fatal("expected bullet_blast event")
	}
	if !foundDamageApplied {
		t.Fatal("expected damage_applied event")
	}
	if score := snapshot.PlayerSessions[playerID].Score; score != constants.BaseScore {
		t.Fatalf("expected player score %d, got %d", constants.BaseScore, score)
	}

	drained := scenario.game.DrainPendingPresentationEvents(playerID, pendingEventIDs(snapshot.PendingEvents)...)
	if len(drained) != 2 {
		t.Fatalf("expected 2 drained events, got %d", len(drained))
	}

	flushedSnapshot := scenario.presentationSnapshot(playerID)
	if len(flushedSnapshot.PendingEvents) != 0 {
		t.Fatalf("expected flushed gameplay snapshot event batch to have 0 events, got %d", len(flushedSnapshot.PendingEvents))
	}

	scenario.step(constants.CollisionDespawnDelay)
	snapshot = scenario.presentationSnapshot(playerID)
	if _, ok := snapshot.Bullets["bullet-1"]; ok {
		t.Fatal("expected hit bullet to be removed after despawn delay")
	}
	if _, ok := snapshot.Asteroids["asteroid-1"]; ok {
		t.Fatal("expected hit asteroid to be removed after despawn delay")
	}
}

func TestBulletAsteroidCollisionsSplitLargerAsteroid(t *testing.T) {
	scenario := newScenario(t)
	scenario.useBulletCapsuleAsteroidPolygonCollisions()
	playerID := scenario.addPlayer()
	impactPosition := physics.Vector2{X: 100, Y: 100}
	scenario.placeBullet("bullet-1", playerID, impactPosition, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", impactPosition, 3)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	asteroids := scenario.presentationSnapshot(playerID).Asteroids
	if len(asteroids) != 3 {
		t.Fatalf("expected hit asteroid plus 2 fragments, got %d asteroids", len(asteroids))
	}

	fragmentCount := 0
	for asteroidID, asteroid := range asteroids {
		if asteroidID == "asteroid-1" {
			continue
		}

		fragmentCount++
		if asteroid.Size != 2 {
			t.Fatalf("expected fragment size 2, got %d", asteroid.Size)
		}
		if asteroid.X != impactPosition.X || asteroid.Y != impactPosition.Y {
			t.Fatalf("expected fragment at impact position, got (%v, %v)", asteroid.X, asteroid.Y)
		}
		if scenario.asteroidPendingDespawn(asteroidID) {
			t.Fatal("expected fragment to remain active")
		}
	}

	if fragmentCount != 2 {
		t.Fatalf("expected 2 fragments, got %d", fragmentCount)
	}
}

func TestBulletAsteroidCollisionNonfatalDamageDoesNotDestroyScoreOrFragment(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	impactPosition := physics.Vector2{X: 100, Y: 100}
	scenario.placeBullet("bullet-1", playerID, impactPosition, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", impactPosition, 3)
	scenario.setAsteroidHealth("asteroid-1", 3)
	initialHealth := scenario.asteroidHealth("asteroid-1")

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if !scenario.bulletPendingDespawn("bullet-1") {
		t.Fatal("expected hit bullet to be marked for delayed despawn")
	}
	if !scenario.asteroidExists("asteroid-1") {
		t.Fatal("expected asteroid to remain active after nonfatal hit")
	}
	if scenario.asteroidPendingDespawn("asteroid-1") {
		t.Fatal("expected nonfatal-hit asteroid not to be pending despawn")
	}
	if scenario.asteroidHealth("asteroid-1") >= initialHealth {
		t.Fatalf("expected asteroid health to be reduced from %d, got %d", initialHealth, scenario.asteroidHealth("asteroid-1"))
	}
	if score := scenario.playerSessionState(playerID, playerID).Score; score != 0 {
		t.Fatalf("expected no score for non-destroying hit, got %d", score)
	}
	if len(scenario.presentationSnapshot(playerID).Asteroids) != 1 {
		t.Fatalf("expected no spawned fragments for non-destroying hit, got %d asteroids", len(scenario.presentationSnapshot(playerID).Asteroids))
	}
}

func TestBulletAsteroidCollisionAppliesAsteroidDamageModifiers(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	impactPosition := physics.Vector2{X: 100, Y: 100}
	scenario.placeBullet("bullet-1", playerID, impactPosition, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", impactPosition, 3)
	scenario.setAsteroidHealth("asteroid-1", 3)
	scenario.setAsteroidDamageModifiers("asteroid-1", []damage.DamageModifier{
		{Type: damage.DamageTypeKinetic, Category: damage.DamageModifierCategoryResistance, Operation: damage.DamageModifierOperationMultiply, Value: 0.5},
	})

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if health := scenario.asteroidHealth("asteroid-1"); health != 2 {
		t.Fatalf("expected asteroid health 2 after resistance, got %d", health)
	}
}

func TestBulletAsteroidCollisionEmitsDamageAppliedEvent(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	impactPosition := physics.Vector2{X: 100, Y: 100}
	scenario.placeBullet("bullet-1", playerID, impactPosition, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", impactPosition, 3)
	scenario.setAsteroidHealth("asteroid-1", 3)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	events := scenario.pendingPresentationEvents(playerID)
	found := false
	for _, pending := range events {
		event := pending.Event
		if event.Type == "damage_applied" {
			found = true
			if event.SourceType != "projectile" {
				t.Fatalf("expected source type projectile, got %q", event.SourceType)
			}
			if event.SourceID != "bullet-1" {
				t.Fatalf("expected source id bullet-1, got %q", event.SourceID)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected damage_applied event for bullet asteroid collision")
	}
}

func TestBulletAsteroidCollisionsScoreByAsteroidSize(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	impactPosition := physics.Vector2{X: 100, Y: 100}
	scenario.placeBullet("bullet-1", playerID, impactPosition, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", impactPosition, 3)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	expectedScore := constants.BaseScore / 3
	if score := scenario.playerSessionState(playerID, playerID).Score; score != expectedScore {
		t.Fatalf("expected player score %d, got %d", expectedScore, score)
	}
}

func TestBulletAsteroidCollisionWorksAcrossHorizontalBoundary(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.placeBullet("bullet-1", playerID, physics.Vector2{X: 5, Y: 100}, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: constants.WorldWidth - 5, Y: 100}, 3)

	scenario.step(0)

	if !scenario.bulletPendingDespawn("bullet-1") {
		t.Fatal("expected cross-boundary bullet to collide with asteroid")
	}
	if !scenario.asteroidPendingDespawn("asteroid-1") {
		t.Fatal("expected cross-boundary asteroid to be hit by bullet")
	}
}

func TestBulletAsteroidCollisionWorksAcrossVerticalBoundary(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.placeBullet("bullet-1", playerID, physics.Vector2{X: 100, Y: 5}, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: 100, Y: constants.WorldHeight - 5}, 3)

	scenario.step(0)

	if !scenario.bulletPendingDespawn("bullet-1") {
		t.Fatal("expected cross-boundary bullet to collide with asteroid")
	}
	if !scenario.asteroidPendingDespawn("asteroid-1") {
		t.Fatal("expected cross-boundary asteroid to be hit by bullet")
	}
}

func TestPausedPlayerDoesNotScoreFromBulletAsteroidCollision(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	impactPosition := physics.Vector2{X: 100, Y: 100}
	scenario.setPlayerPaused(playerID, true)
	scenario.placeBullet("bullet-1", playerID, impactPosition, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", impactPosition, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if score := scenario.playerSessionState(playerID, playerID).Score; score != 0 {
		t.Fatalf("expected paused player score to remain 0, got %d", score)
	}
}

func TestInvulnerablePlayerDoesNotScoreFromBulletAsteroidCollision(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	impactPosition := physics.Vector2{X: 100, Y: 100}
	scenario.setPlayerInvulnerability(playerID, 1)
	scenario.placeBullet("bullet-1", playerID, impactPosition, physics.Vector2{})
	scenario.placeAsteroid("asteroid-1", impactPosition, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if score := scenario.playerSessionState(playerID, playerID).Score; score != 0 {
		t.Fatalf("expected invulnerable player score to remain 0, got %d", score)
	}
}

func TestShipAsteroidCollisionsDelayPlayerRemovalAndBroadcastDeath(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	otherPlayerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	position := physics.Vector2{X: player.X, Y: player.Y}
	scenario.placeAsteroid("asteroid-1", position, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if !scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected hit player to be marked for delayed despawn")
	}
	if scenario.asteroidPendingDespawn("asteroid-1") {
		t.Fatal("expected ship collision to leave asteroid active")
	}
	if !scenario.playerEntityExists(otherPlayerID) {
		t.Fatal("expected untouched player to remain")
	}

	playerSnapshot := scenario.presentationSnapshot(playerID)
	for _, viewerID := range []string{playerID, otherPlayerID} {
		if events := scenario.pendingEventCount(viewerID); events != 2 {
			t.Fatalf("expected 2 queued events for %s, got %d", viewerID, events)
		}

		snapshot := scenario.presentationSnapshot(viewerID)
		if len(snapshot.PendingEvents) != 2 {
			t.Fatalf("expected 2 events in gameplay snapshot event batch for %s, got %d", viewerID, len(snapshot.PendingEvents))
		}
		foundShipDeath := false
		foundDamageApplied := false
		for _, pending := range snapshot.PendingEvents {
			event := pending.Event
			switch event.Type {
			case servergame.PacketTypeShipDeath:
				foundShipDeath = true
				if event.PlayerID != playerID {
					t.Fatalf("expected dead player id %s for %s, got %q", playerID, viewerID, event.PlayerID)
				}
				if event.X != position.X || event.Y != position.Y {
					t.Fatalf("expected death event at player position for %s, got (%v, %v)", viewerID, event.X, event.Y)
				}
			case "damage_applied":
				foundDamageApplied = true
			}
		}
		if !foundShipDeath {
			t.Fatalf("expected ship_death event for %s", viewerID)
		}
		if !foundDamageApplied {
			t.Fatalf("expected damage_applied event for %s", viewerID)
		}
	}

	drained := scenario.game.DrainPendingPresentationEvents(playerID, pendingEventIDs(playerSnapshot.PendingEvents)...)
	if len(drained) != 2 {
		t.Fatalf("expected 2 drained events, got %d", len(drained))
	}

	if flushedSnapshot := scenario.presentationSnapshot(playerID); len(flushedSnapshot.PendingEvents) != 0 {
		t.Fatalf("expected flushed gameplay snapshot event batch to have 0 events, got %d", len(flushedSnapshot.PendingEvents))
	}

	scenario.step(constants.CollisionDespawnDelay)

	if scenario.playerEntityExists(playerID) {
		t.Fatal("expected hit player to be removed after despawn delay")
	}
	if !scenario.playerEntityExists(otherPlayerID) {
		t.Fatal("expected untouched player to remain after despawn delay")
	}
}

func TestShipAsteroidCollisionWorksAcrossHorizontalBoundary(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.setPlayerPosition(playerID, physics.Vector2{X: 5, Y: 100})
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: constants.WorldWidth - 5, Y: 100}, 1)

	scenario.step(0)

	if !scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected cross-boundary asteroid to collide with player")
	}
}

func TestShipAsteroidCollisionWorksAcrossVerticalBoundary(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	scenario.setPlayerPosition(playerID, physics.Vector2{X: 100, Y: 5})
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: 100, Y: constants.WorldHeight - 5}, 1)

	scenario.step(0)

	if !scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected cross-boundary asteroid to collide with player")
	}
}

func TestShipAsteroidCollisionNonfatalDamageReducesHealthWithoutDeath(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	position := physics.Vector2{X: player.X, Y: player.Y}
	scenario.setPlayerHealth(playerID, 3)
	initialHealth := scenario.playerHealth(playerID)
	initialLives := scenario.playerSessionState(playerID, playerID).Lives
	scenario.placeAsteroid("asteroid-1", position, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected nonfatal collision player not to be pending despawn")
	}
	if !scenario.playerEntityExists(playerID) {
		t.Fatal("expected nonfatal collision player to remain active")
	}
	if scenario.playerHealth(playerID) >= initialHealth {
		t.Fatalf("expected player health to be reduced from %d, got %d", initialHealth, scenario.playerHealth(playerID))
	}
	events := scenario.pendingPresentationEvents(playerID)
	foundShipDeath := false
	foundDamageApplied := false
	for _, pending := range events {
		event := pending.Event
		switch event.Type {
		case servergame.PacketTypeShipDeath:
			foundShipDeath = true
		case "damage_applied":
			foundDamageApplied = true
		}
	}
	if foundShipDeath {
		t.Fatal("expected no ship_death event for nonfatal collision")
	}
	if !foundDamageApplied {
		t.Fatal("expected damage_applied event for nonfatal collision")
	}
	if lives := scenario.playerSessionState(playerID, playerID).Lives; lives != initialLives {
		t.Fatalf("expected lives to remain %d after nonfatal collision, got %d", initialLives, lives)
	}
}

func TestShipAsteroidCollisionAppliesPlayerDamageModifiers(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	position := physics.Vector2{X: player.X, Y: player.Y}
	scenario.setPlayerHealth(playerID, 3)
	scenario.setPlayerDamageModifiers(playerID, []damage.DamageModifier{
		{Type: damage.DamageTypeKinetic, Category: damage.DamageModifierCategoryVulnerability, Operation: damage.DamageModifierOperationMultiply, Value: 1.5},
	})
	scenario.placeAsteroid("asteroid-1", position, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if health := scenario.playerHealth(playerID); health != 1 {
		t.Fatalf("expected player health 1 after vulnerability, got %d", health)
	}
	if scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected player to remain active after modified nonfatal collision")
	}
}

func TestShipAsteroidCollisionEmitsDamageAppliedEvent(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	position := physics.Vector2{X: player.X, Y: player.Y}
	scenario.setPlayerHealth(playerID, 3)
	scenario.placeAsteroid("asteroid-1", position, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	events := scenario.pendingPresentationEvents(playerID)
	found := false
	for _, pending := range events {
		event := pending.Event
		if event.Type == "damage_applied" {
			found = true
			if event.SourceType != "asteroid" {
				t.Fatalf("expected source type asteroid, got %q", event.SourceType)
			}
			break
		}
	}
	if !found {
		t.Fatal("expected damage_applied event for asteroid player collision")
	}
}

func TestShipAsteroidCollisionSkipsPausedPlayer(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.send(playerID, servergame.ClientPacket{Type: servergame.PacketTypePauseRequest})
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected paused player to ignore asteroid collision")
	}
	if events := scenario.pendingEventCount(playerID); events != 0 {
		t.Fatalf("expected no death event for paused player, got %d", events)
	}
}

func TestShipAsteroidCollisionSkipsInvulnerablePlayer(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.setPlayerInvulnerability(playerID, 1)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	scenario.step(1.0 / float64(constants.ServerTickRate))

	if scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected invulnerable player to ignore asteroid collision")
	}
	if events := scenario.pendingEventCount(playerID); events != 0 {
		t.Fatalf("expected no death event for invulnerable player, got %d", events)
	}
}

func TestShipAsteroidCollisionKillsAfterInvulnerabilityExpires(t *testing.T) {
	scenario := newScenario(t)
	scenario.useCircleCollisionShapes()
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)
	scenario.setPlayerInvulnerability(playerID, 0.1)
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: player.X, Y: player.Y}, 1)

	scenario.step(0.1)

	if !scenario.playerPendingDespawn(playerID) {
		t.Fatal("expected player to die after invulnerability expires")
	}
	if events := scenario.pendingEventCount(playerID); events != 2 {
		t.Fatalf("expected two events after invulnerability expires, got %d", events)
	}
	var foundShipDeath, foundDamageApplied bool
	for _, pending := range scenario.pendingPresentationEvents(playerID) {
		event := pending.Event
		switch event.Type {
		case servergame.PacketTypeShipDeath:
			foundShipDeath = true
		case "damage_applied":
			foundDamageApplied = true
		}
	}
	if !foundShipDeath {
		t.Fatal("expected ship_death event")
	}
	if !foundDamageApplied {
		t.Fatal("expected damage_applied event")
	}
}

func TestAsteroidVisibilityUsesCameraViewsWithoutPlayer(t *testing.T) {
	scenario := newScenario(t)
	scenario.addCameraView("player-1", physics.Vector2{X: 100, Y: 100}, runtime.ClientConfig{
		VisibleWorldWidth:  200,
		VisibleWorldHeight: 200,
	})
	scenario.placeAsteroid("asteroid-1", physics.Vector2{X: 100, Y: 100}, 1)

	scenario.step(0)

	if !scenario.asteroidExists("asteroid-1") {
		t.Fatal("expected asteroid inside camera view to remain even without a player entity")
	}
}

