package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func (session *webSocketSession) handleCreateRoomRequest() {
	if !requireAuthenticatedAccount(session) {
		return
	}

	if session.currentRoomID != "" {
		session.EnqueueRoomError(rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, err := session.rooms.CreateLobbyRoom()
	if err != nil {
		logging.Rooms.Error("create lobby room failed", err, "session_id", session.sessionID)
		session.EnqueueRoomError(rooms.RoomErrorInvalidRoomState, "Could not create room.")
		return
	}

	addSessionMember(room, session.sessionID, session)
	session.room = room
	session.currentRoomID = room.ID
	session.currentGamePlayerID = ""
	session.resetDebugShapeCatalogSent()
	session.EnqueueRoomSnapshot(room)
}

func (session *webSocketSession) handleJoinRoomRequest(roomCode string) {
	if !requireAuthenticatedAccount(session) {
		return
	}

	if session.currentRoomID != "" {
		session.EnqueueRoomError(rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, roomErr := session.rooms.JoinRoom(session.sessionID, roomCode)
	if roomErr != nil {
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}

	attachRoomSession(room, session.sessionID, session)
	if accountID := accountIDForSession(session); accountID != "" {
		room.SetMemberAccountIDForSession(session.sessionID, accountID)
	}
	session.room = room
	session.currentRoomID = room.ID
	session.currentGamePlayerID = ""
	session.resetDebugShapeCatalogSent()
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleLeaveRoomRequest() {
	session.leaveRequestedRoom()
}

func (session *webSocketSession) handleSetReadyRequest(ready bool) {
	if session.currentRoomID == "" || session.sessionID == "" {
		session.EnqueueRoomError(rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.SetReady(session.currentRoomID, session.sessionID, ready)
	if roomErr != nil {
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}

	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleStartGameRequest() {
	if session.room == nil || session.sessionID == "" {
		session.EnqueueRoomError(rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.StartRoomGame(session.currentRoomID, session.sessionID)
	if roomErr != nil {
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}

	session.room = room
	session.resetDebugShapeCatalogSent()
	activateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleStartSinglePlayerRequest(localProfileID string) {
	_ = localProfileID
	logging.Network.Debug("StartSinglePlayerRequest received",
		logging.FieldRoomID, session.currentRoomID,
		logging.FieldPlayerID, session.currentGamePlayerID,
		"session_id", session.sessionID,
		"current_room_id", session.currentRoomID,
	)

	if session.currentRoomID != "" {
		session.EnqueueRoomError(rooms.RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, roomErr := session.rooms.CreateStartedSinglePlayerRoom(session.sessionID)
	if roomErr != nil {
		logging.Rooms.Error("create single-player room failed", roomErr, "session_id", session.sessionID)
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}

	attachRoomSession(room, session.sessionID, session)
	session.room = room
	session.currentRoomID = room.ID
	session.currentGamePlayerID = ""
	session.resetDebugShapeCatalogSent()
	if localProfileID != "" {
		room.SetMemberLocalProfileIDForSession(session.sessionID, localProfileID)
	}

	activateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleReturnToLobbyRequest() {
	if session.room == nil || session.sessionID == "" {
		session.EnqueueRoomError(rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.ReturnRoomToLobby(session.currentRoomID, session.sessionID)
	if roomErr != nil {
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}

	session.room = room
	session.resetDebugShapeCatalogSent()
	deactivateRoomPlayers(room)
	BroadcastRoomSnapshot(room)
}
