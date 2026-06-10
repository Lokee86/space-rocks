package networking

import (
	"net/http"

	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

func WebSocketHandler(roomManager *rooms.RoomManager) http.HandlerFunc {
	return WebSocketHandlerWithAuth(roomManager, nil)
}

func WebSocketHandlerWithAuth(roomManager *rooms.RoomManager, verifier TokenVerifier) http.HandlerFunc {
	return WebSocketHandlerWithAuthAndReporter(roomManager, verifier, rooms.NoopMatchResultReporter{})
}

func WebSocketHandlerWithAuthAndReporter(roomManager *rooms.RoomManager, verifier TokenVerifier, reporter rooms.MatchResultReporter) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return allowWebSocketOrigin(r)
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Network.Error("websocket upgrade failed", err, logging.FieldRemoteAddr, r.RemoteAddr)
			return
		}

		session := newWebSocketSession(conn, roomManager, verifier, reporter)
		handleConnection(session, r.RemoteAddr)
	}
}

func handleConnection(session *webSocketSession, remoteAddr string) {
	defer session.conn.Close()
	defer session.leaveDisconnectedRoom()

	roomID := session.currentRoomID
	playerID := session.currentGamePlayerID
	logging.Network.Debug("websocket connected",
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		"session_id", session.sessionID,
		"current_room_id", session.currentRoomID,
		logging.FieldRemoteAddr, remoteAddr,
	)
	defer logging.Network.Debug("websocket disconnected",
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		"session_id", session.sessionID,
		logging.FieldRemoteAddr, remoteAddr,
	)

	readErr := make(chan error, 1)
	gameplayLifecycleDone := make(chan struct{})
	defer close(gameplayLifecycleDone)

	go readClientInput(session, remoteAddr, readErr)
	go tickSessionGameplayLifecycle(session, gameplayLifecycleDone)

	writeServerMessages(session, remoteAddr, readErr)
}

func (session *webSocketSession) leaveRequestedRoom() {
	if session.currentRoomID == "" || session.room == nil {
		session.EnqueueRoomError(rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room := session.room
	roomID := session.currentRoomID
	sessionID := session.sessionID
	playerID := session.currentGamePlayerID

	leaveResult, roomErr := session.rooms.LeaveMember(roomID, sessionID, playerID)
	if roomErr == nil {
		room = leaveResult.Room
		logging.Rooms.Debug("room member left",
			logging.FieldRoomID, roomID,
			"session_id", sessionID,
			"remaining_members", leaveResult.RemainingMembers,
		)
	}
	if sessionID != "" {
		detachRoomSession(room, sessionID)
	}

	session.room = nil
	session.currentRoomID = ""
	session.currentGamePlayerID = ""

	if room.MemberCount() > 0 {
		logging.Rooms.Debug("broadcasting room snapshot after member left",
			logging.FieldRoomID, roomID,
			"session_id", sessionID,
			"remaining_members", room.MemberCount(),
		)
		BroadcastRoomSnapshot(room)
	}
}

func (session *webSocketSession) leaveDisconnectedRoom() {
	if session.currentRoomID == "" || session.room == nil {
		return
	}

	room := session.room
	roomID := session.currentRoomID
	sessionID := session.sessionID
	playerID := session.currentGamePlayerID

	leaveResult, roomErr := session.rooms.LeaveMember(roomID, sessionID, playerID)
	if roomErr == nil {
		room = leaveResult.Room
		logging.Rooms.Debug("room member left",
			logging.FieldRoomID, roomID,
			"session_id", sessionID,
			"remaining_members", leaveResult.RemainingMembers,
		)
	}
	if sessionID != "" {
		detachRoomSession(room, sessionID)
	}

	session.room = nil
	session.currentRoomID = ""
	session.currentGamePlayerID = ""

	if room.MemberCount() > 0 {
		logging.Rooms.Debug("broadcasting room snapshot after member left",
			logging.FieldRoomID, roomID,
			"session_id", sessionID,
			"remaining_members", room.MemberCount(),
		)
		BroadcastRoomSnapshot(room)
	}
}
