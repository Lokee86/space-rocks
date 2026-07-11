package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/weapons"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
)

func TestControlForceRespawnPlayerCreatesCameraViewWithDummyConfig(t *testing.T) {
	gameInstance := New()
	control := NewControl(gameInstance)
	playerID := "player-1"
	spawnPosition := physics.Vector2{X: 320, Y: 420}

	session := newPlayerSession(playerID, spawnPosition)
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
	gameInstance.playerSessions[playerID] = session

	delete(gameInstance.cameraViews, playerID)

	if !control.ForceRespawnPlayer(playerID, spawnPosition, DummyPlayerCameraConfig()) {
		t.Fatal("expected ForceRespawnPlayer to succeed")
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
