package networkingtests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TestRoomLifecycleStateNames(t *testing.T) {
	tests := map[rooms.RoomState]string{
		rooms.RoomStateLobby:    "Lobby",
		rooms.RoomStateStarting: "Starting",
		rooms.RoomStateInGame:   "InGame",
		rooms.RoomStateGameOver: "GameOver",
		rooms.RoomStateClosed:   "Closed",
	}

	for state, expected := range tests {
		if string(state) != expected {
			t.Fatalf("expected room state %q, got %q", expected, state)
		}
	}
}
