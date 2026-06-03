package rooms

import (
	"testing"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

func TestNewRoomInitializesGameInstance(t *testing.T) {
	gameInstance := game.New()
	room := NewRoom("room", RoomStateLobby, gameInstance)

	if got := room.GameInstance(); got == nil {
		t.Fatal("expected room to initialize with a game instance")
	}
	if got := room.GameInstance(); got != gameInstance {
		t.Fatal("expected room to retain the provided game instance")
	}
}

func TestRoomGameInstanceSetAndClear(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)
	gameInstance := game.New()

	room.SetGameInstance(gameInstance)
	if got := room.GameInstance(); got != gameInstance {
		t.Fatal("expected SetGameInstance to update the room game instance")
	}

	room.ClearGameInstance()
	if got := room.GameInstance(); got != nil {
		t.Fatal("expected ClearGameInstance to clear the room game instance")
	}
}

func TestRoomActivePlayerCountSet(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	room.SetActivePlayerCount(3)
	if got := room.ActivePlayerCount(); got != 3 {
		t.Fatalf("expected active player count 3, got %d", got)
	}
}

func TestRoomCleanupVersionMethods(t *testing.T) {
	room := NewRoom("room", RoomStateLobby, nil)

	if got := room.CurrentCleanupVersion(); got != 0 {
		t.Fatalf("expected initial cleanup version 0, got %d", got)
	}
	if room.CleanupVersionMatches(1) {
		t.Fatal("expected version 1 not to match initial cleanup version")
	}

	if got := room.ScheduleCleanupTimer(time.Hour, func(int) {}); got != 1 {
		t.Fatalf("expected first cleanup version to be 1, got %d", got)
	}
	if got := room.CurrentCleanupVersion(); got != 1 {
		t.Fatalf("expected cleanup version 1, got %d", got)
	}
	if !room.CleanupVersionMatches(1) {
		t.Fatal("expected cleanup version 1 to match")
	}

	if got := room.ScheduleCleanupTimer(time.Hour, func(int) {}); got != 2 {
		t.Fatalf("expected second cleanup version to be 2, got %d", got)
	}
	if got := room.CurrentCleanupVersion(); got != 2 {
		t.Fatalf("expected cleanup version 2, got %d", got)
	}
	if !room.CleanupVersionMatches(2) {
		t.Fatal("expected cleanup version 2 to match")
	}
}
