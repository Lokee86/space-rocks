package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
)

func TestControlForceRespawnPlayerCreatesCameraViewWithDummyConfig(t *testing.T) {
	gameInstance := New()
	control := NewControl(gameInstance)
	playerID := gameInstance.AddPlayer()
	spawnPosition := physics.Vector2{X: 320, Y: 420}

	session := gameInstance.playerSessions[playerID]
	session.PlayerArmory = weapons.PlayerArmory{
		Primary: weapons.Equipped{
			ID:         weapons.BasicCannon,
			AmmoPolicy: weapons.InfiniteAmmo,
		},
		Secondary: weapons.Equipped{
			ID:         "auxiliary",
			AmmoPolicy: weapons.LimitedAmmo,
		},
	}
	session.Config = runtime.ClientConfig{
		VisibleWorldWidth:  640,
		VisibleWorldHeight: 360,
	}
	delete(gameInstance.cameraViews, playerID)

	if !control.ForceRespawnPlayer(playerID, spawnPosition, DummyPlayerCameraConfig()) {
		t.Fatal("expected ForceRespawnPlayer to succeed")
	}
	if status, ok := gameInstance.lifeRuntime.Status(playerID); !ok || status != playerstate.StatusActive {
		t.Fatalf("expected devtool force transition to set status active: %q, ok=%t", status, ok)
	}

	cameraView := gameInstance.cameraViews[playerID]
	if cameraView == nil {
		t.Fatalf("expected camera view %q to exist", playerID)
	}
	if cameraView.X != spawnPosition.X || cameraView.Y != spawnPosition.Y {
		t.Fatalf("expected camera position %v, got (%v, %v)", spawnPosition, cameraView.X, cameraView.Y)
	}
	if cameraView.Config.VisibleWorldWidth != DummyPlayerVisibleWorldWidth {
		t.Fatalf("expected camera width %d, got %v", DummyPlayerVisibleWorldWidth, cameraView.Config.VisibleWorldWidth)
	}
	if cameraView.Config.VisibleWorldHeight != DummyPlayerVisibleWorldHeight {
		t.Fatalf("expected camera height %d, got %v", DummyPlayerVisibleWorldHeight, cameraView.Config.VisibleWorldHeight)
	}
}

func TestNewShipCopiesPlayerArmoryIntoShipWeapons(t *testing.T) {
	session := newPlayerSession("player-1", physics.Vector2{X: 100, Y: 200})
	session.PlayerArmory = weapons.PlayerArmory{
		Primary: weapons.Equipped{
			ID:         weapons.BasicCannon,
			AmmoPolicy: weapons.InfiniteAmmo,
		},
		Secondary: weapons.Equipped{
			ID:         "sidearm",
			AmmoPolicy: weapons.LimitedAmmo,
		},
	}

	ship := session.NewShip(physics.Vector2{X: 300, Y: 400})

	if ship.ShipWeapons.Primary != session.PlayerArmory.Primary {
		t.Fatalf("expected primary ship weapons %v, got %v", session.PlayerArmory.Primary, ship.ShipWeapons.Primary)
	}
	if ship.ShipWeapons.Secondary != session.PlayerArmory.Secondary {
		t.Fatalf("expected secondary ship weapons %v, got %v", session.PlayerArmory.Secondary, ship.ShipWeapons.Secondary)
	}
	if ship.WeaponState != (weapons.State{}) {
		t.Fatalf("expected zero weapon state, got %+v", ship.WeaponState)
	}
}

func TestPlayerSessionStatusTransitionsAcrossDeathAndRespawn(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	policy := game.lifeRuntime.Policy()
	if status, ok := game.lifeRuntime.Status(playerID); !ok || status != playerstate.StatusActive {
		t.Fatalf("expected new participant status %q, got %q", playerstate.StatusActive, status)
	}

	deathFact := game.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: playerID})
	if !deathFact.Accepted || deathFact.PlayerID != playerID || deathFact.PreviousStatus != playerstate.StatusActive || deathFact.ResultingStatus != playerstate.StatusPendingRespawn {
		t.Fatalf("unexpected pending death fact: %+v", deathFact)
	}
	if deathFact.RemainingLives != policy.StartingLives-1 || deathFact.RespawnDelay != policy.RespawnDelay || deathFact.DeathCount != 1 {
		t.Fatalf("unexpected pending death fact values: %+v", deathFact)
	}
	if status, ok := game.lifeRuntime.Status(playerID); !ok || status != playerstate.StatusPendingRespawn {
		t.Fatalf("expected death with lives remaining to set status %q, got %q", playerstate.StatusPendingRespawn, status)
	}

	stateBeforeRejectedDeath, ok := game.lifeRuntime.ParticipantSnapshot(playerID)
	if !ok {
		t.Fatal("expected runtime state before rejected death")
	}
	rejectedDeathFact := game.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: playerID})
	if rejectedDeathFact.Accepted || rejectedDeathFact.ReasonCode != "not_active" {
		t.Fatalf("expected repeated death to be rejected with not_active, got %+v", rejectedDeathFact)
	}
	stateAfterRejectedDeath, ok := game.lifeRuntime.ParticipantSnapshot(playerID)
	if !ok || stateAfterRejectedDeath.RemainingLives != stateBeforeRejectedDeath.RemainingLives || stateAfterRejectedDeath.DeathCount != stateBeforeRejectedDeath.DeathCount {
		t.Fatalf("expected rejected death to preserve lives/deaths, got state=%+v", stateAfterRejectedDeath)
	}

	game.lifeRuntime.Step(policy.RespawnDelay)
	respawnFact := game.lifeRuntime.CommitRespawn(playerID)
	if !respawnFact.Accepted {
		t.Fatalf("expected successful respawn to be accepted, got %+v", respawnFact)
	}
	if status, ok := game.lifeRuntime.Status(playerID); !ok || status != playerstate.StatusActive {
		t.Fatalf("expected successful respawn to set status %q, got %q", playerstate.StatusActive, status)
	}

	game.lifeRuntime.SetLives(playerID, 1)
	finalDeathFact := game.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: playerID})
	if !finalDeathFact.Accepted || finalDeathFact.PreviousStatus != playerstate.StatusActive || finalDeathFact.ResultingStatus != playerstate.StatusEliminated || finalDeathFact.RemainingLives != 0 || finalDeathFact.RespawnDelay != 0 || finalDeathFact.DeathCount != 2 {
		t.Fatalf("unexpected eliminated death fact: %+v", finalDeathFact)
	}
	if status, ok := game.lifeRuntime.Status(playerID); !ok || status != playerstate.StatusEliminated {
		t.Fatalf("expected final death to set status %q, got %q", playerstate.StatusEliminated, status)
	}
}

func TestGameRespawnPlayerAcceptsPendingRespawnAndMarksActive(t *testing.T) {
	game := New()
	playerID := game.AddPlayer()
	session := game.playerSessions[playerID]
	delete(game.entities.Players, playerID)
	game.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: playerID})
	game.lifeRuntime.Step(game.lifeRuntime.Policy().RespawnDelay)

	game.respawnPlayer(playerID)

	if _, ok := game.entities.Players[playerID]; !ok {
		t.Fatal("expected accepted respawn to create player entity")
	}
	if status, ok := game.lifeRuntime.Status(playerID); !ok || status != playerstate.StatusActive {
		t.Fatalf("expected accepted respawn to set status %q, got %q", playerstate.StatusActive, status)
	}
	_ = session
}

func TestGameRespawnPlayerRejectsInvalidTransitionsWithoutMutation(t *testing.T) {
	activeGame := New()
	activePlayerID := activeGame.AddPlayer()
	activePlayer := activeGame.entities.Players[activePlayerID]
	activeGame.respawnPlayer(activePlayerID)
	if activeGame.entities.Players[activePlayerID] != activePlayer {
		t.Fatal("expected active-player respawn rejection to preserve entity and status")
	}
	if status, ok := activeGame.lifeRuntime.Status(activePlayerID); !ok || status != playerstate.StatusActive {
		t.Fatal("expected active-player respawn rejection to preserve entity and status")
	}

	missingGame := New()
	missingGame.respawnPlayer("player-1")
	if len(missingGame.entities.Players) != 0 {
		t.Fatal("expected missing-session respawn rejection to preserve empty entities")
	}

	cases := []struct {
		name  string
		setup func(*Game, string)
	}{
		{name: "eliminated", setup: func(game *Game, playerID string) {
			game.lifeRuntime.SetLives(playerID, 1)
			game.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: playerID})
		}},
		{name: "cooldown", setup: func(game *Game, playerID string) {
			game.lifeRuntime.ApplyDeath(lives.DeathInput{PlayerID: playerID})
		}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			game := New()
			playerID := game.AddPlayer()
			delete(game.entities.Players, playerID)
			test.setup(game, playerID)

			game.respawnPlayer(playerID)

			if _, ok := game.entities.Players[playerID]; ok {
				t.Fatal("expected rejected respawn not to create player entity")
			}
			if state, ok := game.lifeRuntime.ParticipantSnapshot(playerID); !ok || state.Status == playerstate.StatusActive {
				t.Fatalf("expected rejected respawn to preserve non-active runtime state: %+v, ok=%t", state, ok)
			}
		})
	}
}
