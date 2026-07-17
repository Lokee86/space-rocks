package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
)

func (session *webSocketSession) handleCreateRoomRequest(traceID string) {
	if !requireAuthenticatedAccount(session, traceID) {
		return
	}

	context := session.sessionContext()
	if context.RoomID != "" {
		session.EnqueueRoomError(traceID, rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, err := session.rooms.CreateLobbyRoom()
	if err != nil {
		if traceID == "" {
			traceID = uuid.NewString()
		}
		logging.Emit(observability.Request{
			Event: observability.EventNameRoomCreationFailed,
			Context: observability.Context{
				TraceID:   traceID,
				SessionID: session.sessionID,
			},
			Fields: observability.Fields{
				"error_code":   "room_creation_failed",
				"failure_mode": "lobby_room_creation_failed",
			},
		})
		session.EnqueueRoomError(traceID, rooms.RoomErrorInvalidRoomState, "Could not create room.")
		return
	}
	if traceID == "" {
		traceID = uuid.NewString()
	}
	logging.Emit(observability.Request{
		Event: observability.EventNameRoomCreated,
		Context: observability.Context{
			TraceID:   traceID,
			SessionID: session.sessionID,
			RoomID:    room.ID,
		},
		Fields: observability.Fields{"reason_code": "lobby_created"},
	})

	addSessionMember(room, session.sessionID, session)
	session.bindRoom(room)
	session.resetDebugShapeCatalogSent()
	session.EnqueueRoomSnapshot(room)
}

func (session *webSocketSession) handleJoinRoomRequest(roomCode string, traceID string) {
	if !requireAuthenticatedAccount(session, traceID) {
		return
	}

	context := session.sessionContext()
	if context.RoomID != "" {
		session.EnqueueRoomError(traceID, rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, roomErr := session.rooms.JoinRoom(session.sessionID, roomCode)
	if roomErr != nil {
		session.EnqueueRoomError(traceID, roomErr.Code, roomErr.Message)
		return
	}

	attachRoomSession(room, session.sessionID, session)
	if accountID := accountIDForSession(session); accountID != "" {
		room.SetMemberAccountIDForSession(session.sessionID, accountID)
	}
	session.bindRoom(room)
	session.resetDebugShapeCatalogSent()
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleLeaveRoomRequest() {
	session.leaveRequestedRoom()
}

func (session *webSocketSession) handleSetReadyRequest(ready bool) {
	context := session.sessionContext()
	if context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.SetReady(context.RoomID, session.sessionID, ready)
	if roomErr != nil {
		session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		return
	}

	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleStartGameRequest() {
	context := session.sessionContext()
	if context.Room == nil || context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.StartRoomGame(context.RoomID, session.sessionID)
	if roomErr != nil {
		session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		return
	}

	session.resetDebugShapeCatalogSent()
	activateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleStartSinglePlayerRequest(localProfileID string, traceID string) {
	_ = localProfileID
	context := session.sessionContext()
	if traceID == "" {
		traceID = uuid.NewString()
	}

	if context.RoomID != "" {
		session.EnqueueRoomError(traceID, rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, roomErr := session.rooms.CreateStartedSinglePlayerRoom(session.sessionID)
	if roomErr != nil {
		logging.Emit(observability.Request{
			Event: observability.EventNameRoomCreationFailed,
			Context: observability.Context{
				TraceID:   traceID,
				SessionID: session.sessionID,
			},
			Fields: observability.Fields{
				"error_code":   roomErr.Code,
				"failure_mode": "single_player_room_creation_failed",
			},
		})
		session.EnqueueRoomError(traceID, roomErr.Code, roomErr.Message)
		return
	}
	logging.Emit(observability.Request{
		Event: observability.EventNameRoomCreated,
		Context: observability.Context{
			TraceID:   traceID,
			SessionID: session.sessionID,
			RoomID:    room.ID,
		},
		Fields: observability.Fields{"reason_code": "single_player_created"},
	})

	attachRoomSession(room, session.sessionID, session)
	session.bindRoom(room)
	session.resetDebugShapeCatalogSent()
	if localProfileID != "" {
		room.SetMemberLocalProfileIDForSession(session.sessionID, localProfileID)
	}

	activateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleReturnToLobbyRequest() {
	context := session.sessionContext()
	if context.Room == nil || context.RoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError("", rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.ReturnRoomToLobby(context.RoomID, session.sessionID)
	if roomErr != nil {
		session.EnqueueRoomError("", roomErr.Code, roomErr.Message)
		return
	}

	session.resetDebugShapeCatalogSent()
	deactivateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}
