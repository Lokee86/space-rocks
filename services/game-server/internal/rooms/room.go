package rooms

import (
	"sync"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

type Room struct {
	ID             string
	State          RoomState
	Game           *game.Game
	membership     *roomMembership
	Joinable       bool
	ActivePlayers  int
	CleanupTimer   *time.Timer
	CleanupVersion int
	mu             sync.Mutex
}

func NewRoom(roomID string, state RoomState, gameInstance *game.Game) *Room {
	return &Room{
		ID:         roomID,
		State:      state,
		Game:       gameInstance,
		membership: newRoomMembership(),
		Joinable:   true,
	}
}
