package outbound

import (
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func TickRoomGameOver(room *rooms.Room, broadcastRoomSnapshot func(*rooms.Room)) bool {
	if !room.MarkGameOverIfComplete() {
		return false
	}

	logging.Rooms.Debug("room game over detected",
		logging.FieldRoomID, room.ID,
	)
	broadcastRoomSnapshot(room)
	return true
}
