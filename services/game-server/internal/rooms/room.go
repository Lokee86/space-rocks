package rooms

import (
	"sync"

	"github.com/Lokee86/space-rocks/server/internal/game"
)

type Room struct {
	ID             string
	State          RoomState
	match          *roomMatch
	membership     *roomMembership
	Joinable       bool
	cleanup        *roomCleanup
	mu             sync.Mutex
}

func NewRoom(roomID string, state RoomState, gameInstance *game.Game) *Room {
	return &Room{
		ID:         roomID,
		State:      state,
		match:      newRoomMatch(gameInstance),
		membership: newRoomMembership(),
		cleanup:    newRoomCleanup(),
		Joinable:   true,
	}
}
