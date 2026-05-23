package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

const (
	DefaultRoomID        = rooms.DefaultRoomID
	RoomCleanupGraceTime = rooms.RoomCleanupGraceTime
	MaxPlayersPerRoom    = rooms.MaxPlayersPerRoom
)

type RoomManager = rooms.RoomManager

func NewRoomManager() *RoomManager {
	return rooms.NewRoomManager()
}

func NewRoomManagerWithCleanupDelay(cleanupDelay time.Duration) *RoomManager {
	return rooms.NewRoomManagerWithCleanupDelay(cleanupDelay)
}
