package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func (session *webSocketSession) leaveRequestedRoom() {
	if session == nil || session.room == nil || session.currentRoomID == "" {
		if session != nil {
			session.EnqueueRoomError(rooms.RoomErrorNotInRoom, "Session is not in a room.")
		}
		return
	}

	session.leaveRoom("requested_leave", true)
}

func (session *webSocketSession) leaveDisconnectedRoom() {
	session.leaveRoom("disconnected", false)
}

func (session *webSocketSession) leaveRoom(reason string, enqueueRoomError bool) {
	if session == nil || session.room == nil || session.currentRoomID == "" {
		return
	}

	if !enqueueRoomError {
		if _, ok := session.rooms.Find(session.currentRoomID); !ok {
			if session.room != nil {
				detachRoomSession(session.room, session.sessionID)
			}
			session.room = nil
			session.currentRoomID = ""
			session.currentGamePlayerID = ""
			session.resetDebugShapeCatalogSent()
			return
		}
	}

	room := session.room
	if !rooms.ReportResolvedMatchResultOnceForReason(room, session.matchResultReporter, reason) {
		// Match reporting is best-effort; leave continues even if nothing is reported.
	}

	leaveResult, roomErr := session.rooms.LeaveMember(session.currentRoomID, session.sessionID, session.currentGamePlayerID)
	if roomErr != nil {
		if !enqueueRoomError && roomErr.Code == rooms.RoomErrorRoomNotFound {
			if session.room != nil {
				detachRoomSession(session.room, session.sessionID)
			}
			session.room = nil
			session.currentRoomID = ""
			session.currentGamePlayerID = ""
			session.resetDebugShapeCatalogSent()
			return
		}
		if enqueueRoomError {
			session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		} else {
			logging.Rooms.Warn("websocket room leave failed",
				logging.FieldError, roomErr,
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				"session_id", session.sessionID,
				"reason", reason,
			)
		}
		return
	}

	detachRoomSession(room, session.sessionID)
	session.room = nil
	session.currentRoomID = ""
	session.currentGamePlayerID = ""
	session.resetDebugShapeCatalogSent()

	if leaveResult != nil && leaveResult.ShouldBroadcastSnapshot && leaveResult.Room != nil {
		BroadcastRoomSnapshot(leaveResult.Room)
	}
}
