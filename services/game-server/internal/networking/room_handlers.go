package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
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
		logging.Rooms.Error("create lobby room failed", err, "session_id", session.sessionID)
		session.EnqueueRoomError(traceID, rooms.RoomErrorInvalidRoomState, "Could not create room.")
		return
	}

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
	logging.Network.Debug("StartSinglePlayerRequest received",
		logging.FieldRoomID, context.RoomID,
		logging.FieldPlayerID, context.GamePlayerID,
		"session_id", session.sessionID,
		"current_room_id", context.RoomID,
	)

	if context.RoomID != "" {
		session.EnqueueRoomError(traceID, rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, roomErr := session.rooms.CreateStartedSinglePlayerRoom(session.sessionID)
	if roomErr != nil {
		logging.Rooms.Error("create single-player room failed", roomErr, "session_id", session.sessionID)
		session.EnqueueRoomError(traceID, roomErr.Code, roomErr.Message)
		return
	}

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
