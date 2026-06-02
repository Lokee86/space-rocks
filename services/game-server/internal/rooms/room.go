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
	Members        map[string]*RoomMember
	OwnerID        string
	Joinable       bool
	ActivePlayers  int
	CleanupTimer   *time.Timer
	CleanupVersion int
	mu             sync.Mutex
}

func NewRoom(roomID string, state RoomState, gameInstance *game.Game) *Room {
	return &Room{
		ID:       roomID,
		State:    state,
		Game:     gameInstance,
		Members:  make(map[string]*RoomMember),
		Joinable: true,
	}
}
