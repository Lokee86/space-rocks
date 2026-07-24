package networking

import "github.com/Lokee86/space-rocks/services/game-server/internal/rooms"

func (session *webSocketSession) handleAddBotRequest() {
	context := session.sessionContext()
	if context.Room == nil || context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, _, roomErr := session.rooms.AddBot(context.RoomID, session.sessionID)
	if roomErr != nil {
		session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		return
	}
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleRemoveRoomMemberRequest(playerID string) {
	context := session.sessionContext()
	if context.Room == nil || context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, removed, roomErr := session.rooms.RemoveRoomMember(context.RoomID, session.sessionID, playerID)
	if roomErr != nil {
		session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		return
	}

	if !removed.IsBot {
		for _, targetSession := range snapshotRoomSessions(room, []string{removed.SessionID}) {
			if targetSession == nil {
				continue
			}
			targetContext := targetSession.sessionContext()
			targetSession.clearRoomContextIfMatch(targetContext)
			targetSession.EnqueueRoomError("", rooms.RoomErrorRemovedByOwner, "You were removed from the room by its owner.")
		}
		detachRoomSession(room, removed.SessionID)
	}

	BroadcastRoomSnapshot(room)
}
