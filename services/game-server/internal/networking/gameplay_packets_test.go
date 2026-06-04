package networking

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/networking/inbound"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestHandleGameplayPacketRoutesClientConfigToGameHandlePacket(t *testing.T) {
	const playerID = "player-1"

	gameInstance := game.New()
	spawnPosition := physics.Vector2{X: 320, Y: 240}
	if !gameInstance.DevtoolsEnsurePlayerSession(playerID, spawnPosition) {
		t.Fatal("expected DevtoolsEnsurePlayerSession to succeed")
	}
	if !gameInstance.DevtoolsSpawnPlayerShip(playerID, spawnPosition, runtime.ClientConfig{
		VisibleWorldWidth:  1920,
		VisibleWorldHeight: 1080,
	}) {
		t.Fatal("expected DevtoolsSpawnPlayerShip to succeed")
	}

	session := &webSocketSession{
		room:                rooms.NewRoom("room-1", rooms.RoomStateInGame, gameInstance),
		currentGamePlayerID: playerID,
	}

	packet := game.ClientPacket{
		Type: game.PacketTypeClientConfig,
		Config: runtime.ClientConfig{
			VisibleWorldWidth:  640,
			VisibleWorldHeight: 360,
		},
	}

	if !inbound.HandleGameplayPacket(gameplayPacketTestAdapter{session: session}, packet) {
		t.Fatal("expected client_config packet to be handled")
	}

	cameraConfig := cameraViewConfigForPlayer(t, gameInstance, playerID)
	if cameraConfig.VisibleWorldWidth != packet.Config.VisibleWorldWidth {
		t.Fatalf("expected camera width %v, got %v", packet.Config.VisibleWorldWidth, cameraConfig.VisibleWorldWidth)
	}
	if cameraConfig.VisibleWorldHeight != packet.Config.VisibleWorldHeight {
		t.Fatalf("expected camera height %v, got %v", packet.Config.VisibleWorldHeight, cameraConfig.VisibleWorldHeight)
	}

	sessionConfig := playerSessionConfigForPlayer(t, gameInstance, playerID)
	if sessionConfig.VisibleWorldWidth != packet.Config.VisibleWorldWidth {
		t.Fatalf("expected session width %v, got %v", packet.Config.VisibleWorldWidth, sessionConfig.VisibleWorldWidth)
	}
	if sessionConfig.VisibleWorldHeight != packet.Config.VisibleWorldHeight {
		t.Fatalf("expected session height %v, got %v", packet.Config.VisibleWorldHeight, sessionConfig.VisibleWorldHeight)
	}
}

type gameplayPacketTestAdapter struct {
	session *webSocketSession
}

func (a gameplayPacketTestAdapter) CurrentRoom() *rooms.Room {
	return a.session.room
}

func (a gameplayPacketTestAdapter) CurrentGamePlayerID() string {
	return a.session.currentGamePlayerID
}

func (a gameplayPacketTestAdapter) EnqueuePlayerPauseState() {
	a.session.EnqueuePlayerPauseState()
}

func cameraViewConfigForPlayer(t *testing.T, gameInstance *game.Game, playerID string) runtime.ClientConfig {
	t.Helper()

	cameraViews := exportedFieldValue(t, gameInstance, "cameraViews")
	cameraView := mapValueForKey(t, cameraViews, playerID)
	if !cameraView.IsValid() {
		t.Fatalf("expected camera view for %q", playerID)
	}

	return clientConfigFieldValue(t, cameraView)
}

func playerSessionConfigForPlayer(t *testing.T, gameInstance *game.Game, playerID string) runtime.ClientConfig {
	t.Helper()

	playerSessions := exportedFieldValue(t, gameInstance, "playerSessions")
	session := mapValueForKey(t, playerSessions, playerID)
	if !session.IsValid() {
		t.Fatalf("expected player session for %q", playerID)
	}

	return clientConfigFieldValue(t, session)
}

func exportedFieldValue(t *testing.T, target any, fieldName string) reflect.Value {
	t.Helper()

	field := reflect.ValueOf(target).Elem().FieldByName(fieldName)
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
}

func mapValueForKey(t *testing.T, mapValue reflect.Value, key string) reflect.Value {
	t.Helper()

	value := mapValue.MapIndex(reflect.ValueOf(key))
	if !value.IsValid() {
		return reflect.Value{}
	}

	return value
}

func clientConfigFieldValue(t *testing.T, value reflect.Value) runtime.ClientConfig {
	t.Helper()

	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	field := value.FieldByName("Config")
	return reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem().Interface().(runtime.ClientConfig)
}
