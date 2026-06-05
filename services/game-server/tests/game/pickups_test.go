package gametests

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestPickupSpawnStoresPickup(t *testing.T) {
	scenario := newScenario(t)

	spawnedPickup, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: 120, Y: 220})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}
	if spawnedPickup == nil {
		t.Fatal("expected spawned pickup")
	}

	storedPickup := scenario.pickups().MapIndex(reflect.ValueOf(spawnedPickup.ID))
	if !storedPickup.IsValid() || storedPickup.IsNil() {
		t.Fatalf("expected pickup %q to be stored", spawnedPickup.ID)
	}
}

func TestPickupSpawnUsesStableIDAndDefinitionType(t *testing.T) {
	scenario := newScenario(t)

	definition, ok := pickups.DefinitionFor(pickups.TypeOneUp)
	if !ok {
		t.Fatal("expected pickup definition")
	}

	spawnedPickup, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: 10, Y: 20})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}
	if spawnedPickup.ID != "pickup_1" {
		t.Fatalf("expected pickup ID %q, got %q", "pickup_1", spawnedPickup.ID)
	}
	if spawnedPickup.Type != definition.Type {
		t.Fatalf("expected pickup type %q, got %q", definition.Type, spawnedPickup.Type)
	}
}

func TestPickupSpawnInitializesHealthFromDefinition(t *testing.T) {
	scenario := newScenario(t)

	definition, ok := pickups.DefinitionFor(pickups.TypeOneUp)
	if !ok {
		t.Fatal("expected pickup definition")
	}

	spawnedPickup, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: 10, Y: 20})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}
	if spawnedPickup.Health != definition.Health {
		t.Fatalf("expected pickup health %d, got %d", definition.Health, spawnedPickup.Health)
	}
}

func TestPickupSpawnRejectsUnknownType(t *testing.T) {
	scenario := newScenario(t)

	spawnedPickup, ok, err := scenario.game.SpawnPickup(pickups.PickupType("unknown"), physics.Vector2{})
	if err == nil {
		t.Fatal("expected unknown pickup type to be rejected")
	}
	if ok {
		t.Fatal("expected unknown pickup type to return false")
	}
	if spawnedPickup != nil {
		t.Fatalf("expected no pickup for unknown type, got %+v", spawnedPickup)
	}
}

func TestPickupRemoveDeletesExistingPickup(t *testing.T) {
	scenario := newScenario(t)

	spawnedPickup, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: 32, Y: 48})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}

	if removed := scenario.game.RemovePickup(spawnedPickup.ID); !removed {
		t.Fatalf("expected pickup %q to be removed", spawnedPickup.ID)
	}
	if scenario.pickups().MapIndex(reflect.ValueOf(spawnedPickup.ID)).IsValid() {
		t.Fatalf("expected pickup %q to be deleted from store", spawnedPickup.ID)
	}
}

func TestPickupRemoveReturnsFalseForMissingPickup(t *testing.T) {
	scenario := newScenario(t)

	if removed := scenario.game.RemovePickup("pickup_missing"); removed {
		t.Fatal("expected missing pickup removal to return false")
	}
}

func TestStatePacketIncludesSpawnedPickups(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	spawnedPickup, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: 12, Y: 34})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}

	packet := scenario.state(playerID)
	if len(packet.Pickups) != 1 {
		t.Fatalf("expected one pickup in state packet, got %d", len(packet.Pickups))
	}

	pickup, ok := packet.Pickups[spawnedPickup.ID]
	if !ok {
		t.Fatalf("expected state packet to include pickup %q", spawnedPickup.ID)
	}
	if pickup.ID != spawnedPickup.ID {
		t.Fatalf("expected pickup id %q, got %q", spawnedPickup.ID, pickup.ID)
	}
	if pickup.Type != string(spawnedPickup.Type) {
		t.Fatalf("expected pickup type %q, got %q", spawnedPickup.Type, pickup.Type)
	}
	if pickup.X != spawnedPickup.X || pickup.Y != spawnedPickup.Y {
		t.Fatalf("expected pickup at (%v, %v), got (%v, %v)", spawnedPickup.X, spawnedPickup.Y, pickup.X, pickup.Y)
	}
	if pickup.Health != spawnedPickup.Health {
		t.Fatalf("expected pickup health %d, got %d", spawnedPickup.Health, pickup.Health)
	}
}

func TestStatePacketUsesEmptyPickupStateWhenNoPickupsExist(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()

	packet := scenario.state(playerID)
	if packet.Pickups == nil {
		t.Fatal("expected state packet pickups map to be initialized")
	}
	if len(packet.Pickups) != 0 {
		t.Fatalf("expected empty pickup state, got %d pickups", len(packet.Pickups))
	}
}

func TestPickupCollisionRemovesPickup(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)

	_, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: player.X, Y: player.Y})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}

	scenario.useCircleCollisionShapes()
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if pickups := scenario.pickups(); pickups.Len() != 0 {
		t.Fatalf("expected pickup collision to remove pickup, got %d pickups", pickups.Len())
	}
}

func TestPickupCollisionRespectsFreezeCollisions(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)

	_, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: player.X, Y: player.Y})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}

	scenario.useCircleCollisionShapes()
	devtools.HandleCommand(scenario.game, playerID, devtools.DebugCommand{
		Type:         devtools.PacketTypeToggleDebugFreezeWorld,
		FreezeTarget: "collisions",
	})
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if pickups := scenario.pickups(); pickups.Len() != 1 {
		t.Fatalf("expected frozen collisions to keep pickup, got %d pickups", pickups.Len())
	}
}

func TestOneUpPickupIncrementsPlayerSessionLives(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)

	scenario.setPlayerLives(playerID, 3)
	_, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: player.X, Y: player.Y})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}

	scenario.useCircleCollisionShapes()
	scenario.step(1.0 / float64(constants.ServerTickRate))

	if lives := scenario.playerSessionState(playerID, playerID).Lives; lives != 4 {
		t.Fatalf("expected one_up pickup to increment session lives to 4, got %d", lives)
	}
}

func TestOneUpPickupStatePacketReportsNewLives(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)

	scenario.setPlayerLives(playerID, 2)
	_, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: player.X, Y: player.Y})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}

	scenario.useCircleCollisionShapes()
	scenario.step(1.0 / float64(constants.ServerTickRate))

	packet := scenario.state(playerID)
	session, ok := packet.PlayerSessions[playerID]
	if !ok {
		t.Fatalf("expected state packet to include player session %q", playerID)
	}
	if session.Lives != 3 {
		t.Fatalf("expected state packet player session lives 3, got %d", session.Lives)
	}
}

func TestOneUpPickupCollectionEmitsPickupCollectedAndEffectAppliedEvents(t *testing.T) {
	scenario := newScenario(t)
	playerID := scenario.addPlayer()
	player := scenario.playerState(playerID, playerID)

	scenario.setPlayerLives(playerID, 5)
	_, ok, err := scenario.game.SpawnPickup(pickups.TypeOneUp, physics.Vector2{X: player.X, Y: player.Y})
	if err != nil {
		t.Fatalf("expected pickup spawn to succeed, got error %v", err)
	}
	if !ok {
		t.Fatal("expected pickup spawn to return ok")
	}

	scenario.useCircleCollisionShapes()
	scenario.step(1.0 / float64(constants.ServerTickRate))

	events := scenario.state(playerID).Events
	if len(events) != 2 {
		t.Fatalf("expected two pickup events, got %d", len(events))
	}
	collectedEvent := events[0]
	if collectedEvent.Type != "pickup_collected" {
		t.Fatalf("expected first event type %q, got %q", "pickup_collected", collectedEvent.Type)
	}
	if collectedEvent.PlayerID != playerID {
		t.Fatalf("expected pickup_collected player id %q, got %q", playerID, collectedEvent.PlayerID)
	}
	if collectedEvent.PickupID == "" {
		t.Fatal("expected pickup_collected event to include pickup id")
	}
	if collectedEvent.PickupType != string(pickups.TypeOneUp) {
		t.Fatalf("expected pickup_collected pickup type %q, got %q", pickups.TypeOneUp, collectedEvent.PickupType)
	}
	if collectedEvent.X != player.X || collectedEvent.Y != player.Y {
		t.Fatalf("expected pickup_collected position (%v, %v), got (%v, %v)", player.X, player.Y, collectedEvent.X, collectedEvent.Y)
	}

	effectEvent := events[1]
	if effectEvent.Type != "pickup_effect_applied" {
		t.Fatalf("expected second event type %q, got %q", "pickup_effect_applied", effectEvent.Type)
	}
	if effectEvent.PlayerID != playerID {
		t.Fatalf("expected pickup_effect_applied player id %q, got %q", playerID, effectEvent.PlayerID)
	}
	if effectEvent.PickupID == "" {
		t.Fatal("expected pickup_effect_applied event to include pickup id")
	}
	if effectEvent.PickupType != string(pickups.TypeOneUp) {
		t.Fatalf("expected pickup_effect_applied pickup type %q, got %q", pickups.TypeOneUp, effectEvent.PickupType)
	}
	if effectEvent.EffectType != "add_lives" {
		t.Fatalf("expected effect type %q, got %q", "add_lives", effectEvent.EffectType)
	}
	if effectEvent.Amount != 1 {
		t.Fatalf("expected effect amount 1, got %d", effectEvent.Amount)
	}
	if effectEvent.LivesAfter != 6 {
		t.Fatalf("expected lives after 6, got %d", effectEvent.LivesAfter)
	}
}
