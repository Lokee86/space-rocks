package game

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func TestDevtoolsSpawnPlayerShipUsesDummyCameraConfig(t *testing.T) {
	gameInstance := New()
	control := NewControl(gameInstance)
	playerID := "player-1"
	spawnPosition := physics.Vector2{X: 120, Y: 220}

	if !control.EnsurePlayerSession(playerID, spawnPosition) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}

	session := gameInstance.playerSessions[playerID]
	if session == nil {
		t.Fatalf("expected session %q to exist", playerID)
	}
	session.Config = runtime.ClientConfig{
		VisibleWorldWidth:  640,
		VisibleWorldHeight: 360,
	}

	if !control.SpawnPlayerShip(playerID, spawnPosition, DummyPlayerCameraConfig()) {
		t.Fatal("expected SpawnPlayerShip to succeed")
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

func TestDevtoolsTargetPlayerIDsIncludesSessionAndShipTargets(t *testing.T) {
	gameInstance := New()
	control := NewControl(gameInstance)
	sessionOnlyID := "player-2"
	sharedID := "player-3"
	shipOnlyID := "player-4"
	spawnPosition := physics.Vector2{X: 120, Y: 220}

	if !control.EnsurePlayerSession(sessionOnlyID, spawnPosition) {
		t.Fatal("expected EnsurePlayerSession to create session-only target")
	}
	if !control.EnsurePlayerSession(sharedID, spawnPosition) {
		t.Fatal("expected EnsurePlayerSession to create shared target session")
	}
	if !control.SpawnPlayerShip(sharedID, spawnPosition, DummyPlayerCameraConfig()) {
		t.Fatal("expected SpawnPlayerShip to create shared target ship")
	}

	gameInstance.entities.Players[shipOnlyID] = &runtime.Ship{ID: shipOnlyID, X: 10, Y: 20}

	got := control.TargetPlayerIDs()
	want := []string{sessionOnlyID, sharedID, shipOnlyID}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("TargetPlayerIDs() = %v, want %v", got, want)
	}
}
