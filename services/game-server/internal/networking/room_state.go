package networking

import roomdomain "github.com/Lokee86/space-rocks/server/internal/rooms"

type RoomState = roomdomain.RoomState

const (
	RoomStateLobby    = roomdomain.RoomStateLobby
	RoomStateStarting = roomdomain.RoomStateStarting
	RoomStateInGame   = roomdomain.RoomStateInGame
	RoomStateGameOver = roomdomain.RoomStateGameOver
	RoomStateClosed   = roomdomain.RoomStateClosed
)
