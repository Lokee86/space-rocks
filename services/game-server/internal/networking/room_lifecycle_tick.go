package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func checkRoomGameOver(room *rooms.Room) bool {
	if !room.MarkGameOverIfComplete() {
		return false
	}

	logging.Rooms.Debug("room game over detected",
		logging.FieldRoomID, room.ID,
	)
	BroadcastRoomSnapshot(room)
	return true
}
