package rooms

import "github.com/Lokee86/space-rocks/server/internal/logging"

func TickRoomGameOverLifecycle(room *Room, broadcastRoomSnapshot func(*Room)) bool {
	if !room.MarkGameOverIfComplete() {
		return false
	}

	logging.Rooms.Debug("room game over detected",
		logging.FieldRoomID, room.ID,
	)
	broadcastRoomSnapshot(room)
	return true
}
