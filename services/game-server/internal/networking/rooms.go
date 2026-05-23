package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

const (
	DefaultRoomID        = rooms.DefaultRoomID
	RoomCleanupGraceTime = rooms.RoomCleanupGraceTime
)

func NewRoomManager() *rooms.RoomManager {
	return rooms.NewRoomManager()
}

func NewRoomManagerWithCleanupDelay(cleanupDelay time.Duration) *rooms.RoomManager {
	return rooms.NewRoomManagerWithCleanupDelay(cleanupDelay)
}
