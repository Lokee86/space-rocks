package game

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func TestDevtoolsForceRespawnPlayerCreatesCameraViewWithDummyConfig(t *testing.T) {
	game := New()
	playerID := "player-1"
	spawnPosition := physics.Vector2{X: 320, Y: 420}

	session := newPlayerSession(playerID, spawnPosition)
	session.Config = runtime.ClientConfig{
		VisibleWorldWidth:  640,
		VisibleWorldHeight: 360,
	}
	game.playerSessions[playerID] = session

	delete(game.cameraViews, playerID)

	if !game.DevtoolsForceRespawnPlayer(playerID, spawnPosition, DummyPlayerCameraConfig()) {
		t.Fatal("expected DevtoolsForceRespawnPlayer to succeed")
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
