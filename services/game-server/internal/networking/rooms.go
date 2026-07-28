package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

const (
	DefaultRoomID        = rooms.DefaultRoomID
	RoomCleanupGraceTime = rooms.RoomCleanupGraceTime
)

func NewRoomManager() *rooms.RoomManager {
	return rooms.NewRoomManager()
}

func NewRoomManagerWithGameFactory(gameFactory rooms.GameFactory) *rooms.RoomManager {
	return rooms.NewRoomManagerWithGameFactory(gameFactory)
}

func NewRoomManagerWithCleanupDelay(cleanupDelay time.Duration) *rooms.RoomManager {
	return rooms.NewRoomManagerWithCleanupDelay(cleanupDelay)
}
