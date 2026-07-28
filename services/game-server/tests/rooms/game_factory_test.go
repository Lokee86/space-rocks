package roomstests

import (
	"testing"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func TestRoomManagerUsesConfiguredGameFactory(t *testing.T) {
	const seed int64 = 271828
	manager := rooms.NewRoomManagerWithGameFactory(func() *game.Game {
		return game.NewWithSeed(seed)
	})
	room, roomErr := manager.CreateStartedSinglePlayerRoom("session-1")
	if roomErr != nil {
		t.Fatalf("create started room: %v", roomErr)
	}
	if got := room.GameInstance().SimulationSeed(); got != seed {
		t.Fatalf("simulation seed = %d, want %d", got, seed)
	}
}

func TestRoomManagerDefaultsToNormalGameFactory(t *testing.T) {
	manager := rooms.NewRoomManagerWithGameFactory(nil)
	room, roomErr := manager.CreateStartedSinglePlayerRoom("session-1")
	if roomErr != nil {
		t.Fatalf("create started room: %v", roomErr)
	}
	if room.GameInstance() == nil {
		t.Fatal("expected a game instance")
	}
}
