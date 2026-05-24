package networking

import (
	"net/http"

	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

func WebSocketHandler(rooms *rooms.RoomManager) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Network.Error("websocket upgrade failed", err, logging.FieldRemoteAddr, r.RemoteAddr)
			return
		}

		session := newWebSocketSession(conn, rooms)
		handleConnection(session, r.RemoteAddr)
	}
}

func handleConnection(session *webSocketSession, remoteAddr string) {
	defer session.conn.Close()
	defer session.leaveDisconnectedRoom()

	roomID := session.currentRoomID
	playerID := session.currentPlayerID
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
	go readClientInput(session, remoteAddr, readErr)

	writeServerMessages(session, remoteAddr, readErr)
}

func (session *webSocketSession) leaveRequestedRoom() {
	if session.currentRoomID == "" || session.room == nil {
		session.EnqueueRoomError(rooms.RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room := session.room
	roomID := session.currentRoomID
	memberID := session.currentMemberID
	playerID := session.currentPlayerID

	leaveResult, roomErr := session.rooms.LeaveMember(roomID, memberID, playerID)
	if roomErr == nil {
		room = leaveResult.Room
		logging.Rooms.Debug("room member left",
			logging.FieldRoomID, roomID,
			"member_id", memberID,
			"session_id", session.sessionID,
			"remaining_members", leaveResult.RemainingMembers,
		)
	}
	if memberID != "" {
		detachRoomSession(room, memberID)
	}

	session.room = nil
	session.currentRoomID = ""
	session.currentMemberID = ""
	session.currentPlayerID = ""

	if room.MemberCount() > 0 {
		logging.Rooms.Debug("broadcasting room snapshot after member left",
			logging.FieldRoomID, roomID,
			"member_id", memberID,
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
	memberID := session.currentMemberID
	playerID := session.currentPlayerID

	leaveResult, roomErr := session.rooms.LeaveMember(roomID, memberID, playerID)
	if roomErr == nil {
		room = leaveResult.Room
		logging.Rooms.Debug("room member left",
			logging.FieldRoomID, roomID,
			"member_id", memberID,
			"session_id", session.sessionID,
			"remaining_members", leaveResult.RemainingMembers,
		)
	}
	if memberID != "" {
		detachRoomSession(room, memberID)
	}

	session.room = nil
	session.currentRoomID = ""
	session.currentMemberID = ""
	session.currentPlayerID = ""

	if room.MemberCount() > 0 {
		logging.Rooms.Debug("broadcasting room snapshot after member left",
			logging.FieldRoomID, roomID,
			"member_id", memberID,
			"remaining_members", room.MemberCount(),
		)
		BroadcastRoomSnapshot(room)
	}
}
