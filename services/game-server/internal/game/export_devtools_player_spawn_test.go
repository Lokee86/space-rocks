package game

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
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
	session.Config = runtime.ClientConfig{
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

func TestDevtoolsTargetPlayerIDsIncludesSessionAndShipTargets(t *testing.T) {
	game := New()
	sessionOnlyID := "player-2"
	sharedID := "player-3"
	shipOnlyID := "player-4"
	spawnPosition := physics.Vector2{X: 120, Y: 220}

	if !game.DevtoolsEnsurePlayerSession(sessionOnlyID, spawnPosition) {
		t.Fatal("expected DevtoolsEnsurePlayerSession to create session-only target")
	}
	if !game.DevtoolsEnsurePlayerSession(sharedID, spawnPosition) {
		t.Fatal("expected DevtoolsEnsurePlayerSession to create shared target session")
	}
	if !game.DevtoolsSpawnPlayerShip(sharedID, spawnPosition, DummyPlayerCameraConfig()) {
		t.Fatal("expected DevtoolsSpawnPlayerShip to create shared target ship")
	}

	game.state.Players[shipOnlyID] = &runtime.Ship{ID: shipOnlyID, X: 10, Y: 20}

	got := game.DevtoolsTargetPlayerIDs()
	want := []string{sessionOnlyID, sharedID, shipOnlyID}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DevtoolsTargetPlayerIDs() = %v, want %v", got, want)
	}
}
