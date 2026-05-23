package networkingtests

import (
	"testing"

	"github.com/Lokee86/space-rocks/server/internal/networking"
)

func TestRoomLifecycleStateNames(t *testing.T) {
	tests := map[networking.RoomState]string{
		networking.RoomStateLobby:    "Lobby",
		networking.RoomStateStarting: "Starting",
		networking.RoomStateInGame:   "InGame",
		networking.RoomStateGameOver: "GameOver",
		networking.RoomStateClosed:   "Closed",
	}

	for state, expected := range tests {
		if string(state) != expected {
			t.Fatalf("expected room state %q, got %q", expected, state)
		}
	}
}
