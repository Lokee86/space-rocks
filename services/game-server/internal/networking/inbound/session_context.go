package inbound

import "github.com/Lokee86/space-rocks/services/game-server/internal/rooms"

// SessionContext is a snapshot of the session's coherent room state.
type SessionContext struct {
	Room         *rooms.Room
	RoomID       string
	GamePlayerID string
}
