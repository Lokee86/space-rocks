package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/entities"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDevtoolsSpawnPlayerShipUsesDummyCameraConfig(t *testing.T) {
	game := New()
	playerID := "player-1"
	spawnPosition := physics.Vector2{X: 120, Y: 220}

	if !game.DevtoolsEnsurePlayerSession(playerID, spawnPosition) {
		t.Fatal("expected DevtoolsEnsurePlayerSession to succeed")
	}

	session := game.playerSessions[playerID]
	if session == nil {
		t.Fatalf("expected session %q to exist", playerID)
	}
	session.Config = entities.ClientConfig{
		VisibleWorldWidth:  640,
		VisibleWorldHeight: 360,
	}

	if !game.DevtoolsSpawnPlayerShip(playerID, spawnPosition, DummyPlayerCameraConfig()) {
		t.Fatal("expected DevtoolsSpawnPlayerShip to succeed")
	}

	cameraView := game.cameraViews[playerID]
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
