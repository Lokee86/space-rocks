package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
)

func (session *webSocketSession) leaveRequestedRoom() {
	if session == nil {
		return
	}
	context := session.sessionContext()
	if context.Room == nil || context.RoomID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	session.leaveRoom("requested_leave", true)
}

func (session *webSocketSession) leaveDisconnectedRoom() {
	session.leaveRoom("disconnected", false)
}

func (session *webSocketSession) leaveRoom(reason string, enqueueRoomError bool) {
	if session == nil {
		return
	}
	context := session.sessionContext()
	if context.Room == nil || context.RoomID == "" {
		return
	}

	room := context.Room
	roomID := context.RoomID
	playerID := context.GamePlayerID
	if !enqueueRoomError {
		if _, ok := session.rooms.Find(roomID); !ok {
			detachRoomSession(room, session.sessionID)
			session.clearRoomContextIfMatch(context)
			session.resetDebugShapeCatalogSent()
			return
		}
	}

	if !rooms.ReportResolvedMatchResultOnceForReason(room, session.matchResultReporter, reason) {
		// Match reporting is best-effort; leave continues even if nothing is reported.
	}

	leaveResult, roomErr := session.rooms.LeaveMember(roomID, session.sessionID, playerID)
	if roomErr != nil {
		if !enqueueRoomError && roomErr.Code == rooms.RoomErrorRoomNotFound {
			detachRoomSession(room, session.sessionID)
			session.clearRoomContextIfMatch(context)
			session.resetDebugShapeCatalogSent()
			return
		}
		if enqueueRoomError {
			session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		} else {
			traceID := room.CurrentMatchTraceID()
			if traceID == "" {
				traceID = uuid.NewString()
			}
			logging.Emit(observability.Request{
				Event: observability.EventNameRoomLeaveFailed,
				Context: observability.Context{
					TraceID:   traceID,
					SessionID: session.sessionID,
					RoomID:    context.RoomID,
					PlayerID:  context.GamePlayerID,
				},
				Fields: observability.Fields{
					"error_code":   roomErr.Code,
					"failure_mode": reason,
				},
			})
		}
		return
	}

	detachRoomSession(room, session.sessionID)
	session.clearRoomContextIfMatch(context)
	session.resetDebugShapeCatalogSent()

	if leaveResult != nil && leaveResult.ShouldBroadcastSnapshot && leaveResult.Room != nil {
		BroadcastRoomSnapshot(leaveResult.Room)
	}
}
