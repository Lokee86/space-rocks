package rooms

import (
	"sync"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
)

type Room struct {
	ID         string
	State      RoomState
	match      *roomMatch
	membership *roomMembership
	Joinable   bool
	MaxPlayers int
	roomTeams  roomTeams
	cleanup    *roomCleanup
	mu         sync.Mutex
}

func NewRoom(roomID string, state RoomState, gameInstance *game.Game) *Room {
	room, _ := NewRoomWithConfig(roomID, state, gameInstance, DefaultRoomCreationConfig())
	return room
}

func NewRoomWithConfig(roomID string, state RoomState, gameInstance *game.Game, creation RoomCreationConfig) (*Room, error) {
	creation = normalizeRoomCreationConfig(creation)
	if err := validateRoomCreationConfig(creation); err != nil {
		return nil, err
	}
	return &Room{
		ID:         roomID,
		State:      state,
		match:      newRoomMatch(gameInstance),
		membership: newRoomMembership(),
		Joinable:   true,
		MaxPlayers: creation.MaxPlayers,
		roomTeams:  newRoomTeams(creation.TeamConfig),
		cleanup:    newRoomCleanup(),
	}, nil
}
