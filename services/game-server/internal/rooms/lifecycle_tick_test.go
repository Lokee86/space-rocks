package rooms

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestTickRoomGameOverLifecycleTransitionsFinishedGameAndBroadcasts(t *testing.T) {
	finishedGame := game.New()
	markLifecycleTickTestGameOver(t, finishedGame)
	room := NewRoom("room", RoomStateInGame, finishedGame)
	broadcasts := 0

	if !TickRoomGameOverLifecycle(room, func(broadcastRoom *Room) {
		broadcasts++
		if broadcastRoom != room {
			t.Fatal("expected transitioned room to be broadcast")
		}
	}) {
		t.Fatal("expected finished room lifecycle tick to transition")
	}

	if room.State != RoomStateGameOver {
		t.Fatalf("expected room state %q, got %q", RoomStateGameOver, room.State)
	}
	if broadcasts != 1 {
		t.Fatalf("expected 1 broadcast, got %d", broadcasts)
	}
}

func TestTickRoomGameOverLifecycleDoesNotBroadcastWithoutTransition(t *testing.T) {
	activeGame := game.New()
	activeGame.AddPlayer()
	room := NewRoom("room", RoomStateInGame, activeGame)
	broadcasts := 0

	if TickRoomGameOverLifecycle(room, func(*Room) {
		broadcasts++
	}) {
		t.Fatal("expected active room lifecycle tick not to transition")
	}

	if room.State != RoomStateInGame {
		t.Fatalf("expected room state %q, got %q", RoomStateInGame, room.State)
	}
	if broadcasts != 0 {
		t.Fatalf("expected no broadcast, got %d", broadcasts)
	}
}

func markLifecycleTickTestGameOver(t *testing.T, gameInstance *game.Game) {
	t.Helper()

	playerID := gameInstance.AddPlayer()
	value := reflect.ValueOf(gameInstance).Elem()
	session := exportLifecycleTickTestValue(value.FieldByName("playerSessions")).
		MapIndex(reflect.ValueOf(playerID))
	exportLifecycleTickTestValue(session.Elem().FieldByName("Lives")).SetInt(0)
	players := exportLifecycleTickTestValue(value.FieldByName("state").FieldByName("Players"))
	players.SetMapIndex(reflect.ValueOf(playerID), reflect.Value{})
}

func exportLifecycleTickTestValue(value reflect.Value) reflect.Value {
	if value.CanSet() {
		return value
	}

	return reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem()
}
