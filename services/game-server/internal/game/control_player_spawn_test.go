package game

import (
	"reflect"
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func TestControlSpawnPlayerShipUsesDummyCameraConfig(t *testing.T) {
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

func TestControlSpawnPlayerShipKeepsExistingValidCameraConfigWhenSuppliedConfigIsInvalid(t *testing.T) {
	gameInstance := New()
	control := NewControl(gameInstance)
	playerID := "player-1"
	spawnPosition := physics.Vector2{X: 120, Y: 220}

	if !control.EnsurePlayerSession(playerID, spawnPosition) {
		t.Fatal("expected EnsurePlayerSession to succeed")
	}

	gameInstance.cameraViews[playerID] = &runtime.CameraView{
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  800,
			VisibleWorldHeight: 600,
		},
	}

	if !control.SpawnPlayerShip(playerID, spawnPosition, runtime.ClientConfig{}) {
		t.Fatal("expected SpawnPlayerShip to succeed")
	}

	cameraView := gameInstance.cameraViews[playerID]
	if cameraView == nil {
		t.Fatalf("expected camera view %q to exist", playerID)
	}
	if cameraView.Config.VisibleWorldWidth != 800 || cameraView.Config.VisibleWorldHeight != 600 {
		t.Fatalf("expected existing valid camera config to remain unchanged, got %+v", cameraView.Config)
	}
}

func TestControlTargetPlayerIDsIncludesSessionAndShipTargets(t *testing.T) {
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

func TestNormalizeControlPlayerIDValidationAndReservation(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{name: "trim and lowercase", in: "  Player-4  ", want: "player-4", ok: true},
		{name: "malformed prefix", in: "pilot-4", want: "", ok: false},
		{name: "zero rejected", in: "player-0", want: "", ok: false},
		{name: "non numeric rejected", in: "player-x", want: "", ok: false},
		{name: "mixed case rejected", in: "pLaYeR-4", want: "", ok: false},
		{name: "leading zero accepted canonically", in: "player-04", want: "player-4", ok: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := normalizeControlPlayerID(tc.in)
			if ok != tc.ok || got != tc.want {
				t.Fatalf("normalizeControlPlayerID(%q) = (%q, %v), want (%q, %v)", tc.in, got, ok, tc.want, tc.ok)
			}
		})
	}

	gameInstance := New()
	control := NewControl(gameInstance)
	gameInstance.playerSessions["player-9"] = newPlayerSession("player-9", physics.Vector2{})
	gameInstance.entities.Players["player-2"] = &runtime.Ship{ID: "player-2"}
	gameInstance.entities.Players["pLaYeR-6"] = &runtime.Ship{ID: "pLaYeR-6"}

	if !control.PlayerIDOccupied("Player-2") {
		t.Fatal("expected canonical collision for Player-2 and player-2")
	}
	if !control.PlayerIDOccupied("bad-id") {
		t.Fatal("expected malformed requested IDs to be unavailable")
	}
	if control.PlayerIDOccupied("player-4") {
		t.Fatal("expected player-4 to be available")
	}
	if control.PlayerIDOccupied("player-6") {
		t.Fatal("expected invalid existing IDs to not block valid allocation")
	}

	gameInstance.nextID = 3
	if !control.ReservePlayerID("player-4") {
		t.Fatal("expected reservation to succeed")
	}
	if gameInstance.nextID != 4 {
		t.Fatalf("expected nextID to advance to 4, got %d", gameInstance.nextID)
	}
	if got := gameInstance.AddPlayer(); got != "player-5" {
		t.Fatalf("expected AddPlayer after reservation to return player-5, got %q", got)
	}
}
