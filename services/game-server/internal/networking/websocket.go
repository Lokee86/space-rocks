package networking

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

const RoomIDQueryParam = "room_id"

var nextSessionID atomic.Uint64

type webSocketSession struct {
	conn            *websocket.Conn
	sessionID       string
	currentRoomID   string
	currentMemberID string
	currentPlayerID string
	room            *rooms.Room
	rooms           *RoomManager
	outbound        chan []byte
	legacyLeaveRoom func()
}

func WebSocketHandler(rooms *RoomManager) http.HandlerFunc {
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
		roomID := r.URL.Query().Get(RoomIDQueryParam)
		if roomID != "" {
			// Legacy/dev compatibility path. New multiplayer clients connect to
			// /ws only and join rooms with packets after the session exists.
			attachLegacyRoomSession(rooms, session, roomID)
		}
		handleConnection(session, r.RemoteAddr)
	}
}

func newWebSocketSession(conn *websocket.Conn, rooms *RoomManager) *webSocketSession {
	sessionNumber := nextSessionID.Add(1)

	return &webSocketSession{
		conn:      conn,
		sessionID: "session-" + strconv.FormatUint(sessionNumber, 10),
		rooms:     rooms,
		outbound:  make(chan []byte, 16),
	}
}

func attachLegacyRoomSession(rooms *RoomManager, session *webSocketSession, roomID string) {
	room, leaveRoom := rooms.Join(roomID)
	playerID := room.Game.AddPlayer()
	room.AddMemberID(playerID)

	session.room = room
	session.currentRoomID = room.ID
	session.currentMemberID = playerID
	session.currentPlayerID = playerID
	session.legacyLeaveRoom = leaveRoom
}

func handleConnection(session *webSocketSession, remoteAddr string) {
	defer session.conn.Close()
	defer session.leaveCurrentRoom(false)

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

func readClientInput(
	session *webSocketSession,
	remoteAddr string,
	readErr chan<- error,
) {
	for {
		_, msg, err := session.conn.ReadMessage()
		if err != nil {
			readErr <- err
			return
		}

		var packet game.ClientPacket
		if err := json.Unmarshal(msg, &packet); err != nil {
			logging.Network.Warn("websocket packet decode failed",
				logging.FieldError, err,
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentPlayerID,
				"session_id", session.sessionID,
				logging.FieldRemoteAddr, remoteAddr,
			)
			continue
		}

		if packet.Type == game.PacketTypeCreateRoomRequest {
			session.logLobbyPacketReceived("CreateRoomRequest received", "")
			session.handleCreateRoomRequest()
			continue
		}
		if packet.Type == game.PacketTypeJoinRoomRequest {
			session.logLobbyPacketReceived("JoinRoomRequest received", packet.RoomCode)
			session.handleJoinRoomRequest(packet.RoomCode)
			continue
		}
		if packet.Type == game.PacketTypeLeaveRoomRequest {
			session.handleLeaveRoomRequest()
			continue
		}
		if packet.Type == game.PacketTypeSetReadyRequest {
			session.handleSetReadyRequest(packet.Ready)
			continue
		}
		if packet.Type == game.PacketTypeStartGameRequest {
			session.handleStartGameRequest()
			continue
		}

		if session.room == nil || session.currentPlayerID == "" {
			continue
		}

		session.room.Game.HandlePacket(session.currentPlayerID, packet)
	}
}

func (session *webSocketSession) logLobbyPacketReceived(message string, roomCode string) {
	args := []any{
		logging.FieldRoomID, session.currentRoomID,
		logging.FieldPlayerID, session.currentPlayerID,
		"session_id", session.sessionID,
		"current_room_id", session.currentRoomID,
	}
	if roomCode != "" {
		args = append(args, "room_code", roomCode)
	}

	logging.Network.Debug(message, args...)
}

func (session *webSocketSession) handleCreateRoomRequest() {
	if session.currentRoomID != "" {
		session.EnqueueRoomError(RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, err := session.rooms.CreateLobbyRoom()
	if err != nil {
		logging.Rooms.Error("create lobby room failed", err, "session_id", session.sessionID)
		session.EnqueueRoomError(RoomErrorInvalidRoomState, "Could not create room.")
		return
	}

	addSessionMember(room, session.sessionID, session)
	session.room = room
	session.currentRoomID = room.ID
	session.currentMemberID = session.sessionID
	session.currentPlayerID = ""
	session.EnqueueRoomSnapshot(room)
}

func (session *webSocketSession) handleJoinRoomRequest(roomCode string) {
	if session.currentRoomID != "" {
		session.EnqueueRoomError(RoomErrorAlreadyInRoom, "Session is already in a room.")
		return
	}

	room, roomErr := session.rooms.JoinRoom(session.sessionID, roomCode)
	if roomErr != nil {
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}

	attachRoomSession(room, session.sessionID, session)
	session.room = room
	session.currentRoomID = room.ID
	session.currentMemberID = session.sessionID
	session.currentPlayerID = ""
	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleLeaveRoomRequest() {
	session.leaveCurrentRoom(true)
}

func (session *webSocketSession) handleSetReadyRequest(ready bool) {
	if session.currentRoomID == "" || session.currentMemberID == "" {
		session.EnqueueRoomError(RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	room, roomErr := session.rooms.SetReady(session.currentRoomID, session.currentMemberID, ready)
	if roomErr != nil {
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}

	BroadcastRoomSnapshot(room)
}

func (session *webSocketSession) handleStartGameRequest() {
	if session.room == nil || session.currentMemberID == "" {
		session.EnqueueRoomError(RoomErrorNotInRoom, "Session is not in a room.")
		return
	}

	if roomErr := session.room.ValidateStart(session.currentMemberID); roomErr != nil {
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}

	if roomErr := session.room.MarkStarting(); roomErr != nil {
		session.EnqueueRoomError(roomErr.Code, roomErr.Message)
		return
	}
	if session.room.Game == nil {
		session.room.Game = game.New()
	}
	session.room.Game.Start()
	activateRoomPlayers(session.room)
	session.room.State = rooms.RoomStateInGame
	BroadcastRoomSnapshot(session.room)
}

func (session *webSocketSession) leaveCurrentRoom(reportNotInRoom bool) {
	if session.currentRoomID == "" || session.room == nil {
		if reportNotInRoom {
			session.EnqueueRoomError(RoomErrorNotInRoom, "Session is not in a room.")
		}
		return
	}

	room := session.room
	roomID := session.currentRoomID
	memberID := session.currentMemberID
	playerID := session.currentPlayerID

	if memberID != "" {
		if leaveResult, roomErr := session.rooms.LeaveRoom(roomID, memberID); roomErr == nil {
			room = leaveResult.Room
		}
		detachRoomSession(room, memberID)
	}
	if playerID != "" && room.Game != nil {
		room.Game.RemovePlayer(playerID)
		if session.legacyLeaveRoom == nil && room.ActivePlayers > 0 {
			room.ActivePlayers--
		}
		if session.legacyLeaveRoom != nil {
			session.legacyLeaveRoom()
		} else {
			session.rooms.ScheduleCleanupIfEmpty(roomID)
		}
	} else {
		session.rooms.ScheduleCleanupIfEmpty(roomID)
	}

	session.room = nil
	session.currentRoomID = ""
	session.currentMemberID = ""
	session.currentPlayerID = ""
	session.legacyLeaveRoom = nil

	if room.MemberCount() > 0 {
		BroadcastRoomSnapshot(room)
	}
}

func activateRoomPlayers(room *rooms.Room) {
	memberSnapshot := room.MembersSnapshot()
	memberIDs := make([]string, 0, len(memberSnapshot))
	for _, member := range memberSnapshot {
		if !member.Connected {
			continue
		}
		memberIDs = append(memberIDs, member.SessionID)
	}

	sessions := snapshotRoomSessions(room, memberIDs)
	for _, session := range sessions {
		if session == nil || session.currentPlayerID != "" {
			continue
		}

		playerID := room.Game.AddPlayer()
		session.currentPlayerID = playerID
		room.ActivePlayers++
	}
}

func writeServerMessages(
	session *webSocketSession,
	remoteAddr string,
	readErr <-chan error,
) {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	for {
		select {
		case err := <-readErr:
			logWebSocketReadClose(err, session.currentRoomID, session.currentPlayerID, remoteAddr)
			return
		case message := <-session.outbound:
			if err := session.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				logWebSocketWriteClose(err, session.currentRoomID, session.currentPlayerID, remoteAddr)
				return
			}
		case <-ticker.C:
			if session.room == nil || session.currentPlayerID == "" || session.room.State != rooms.RoomStateInGame {
				continue
			}

			response := session.room.Game.State(session.currentPlayerID)
			if response == nil {
				continue
			}

			if err := session.conn.WriteMessage(websocket.TextMessage, response); err != nil {
				logWebSocketWriteClose(err, session.currentRoomID, session.currentPlayerID, remoteAddr)
				return
			}
		}
	}
}

func logWebSocketReadClose(err error, roomID string, playerID string, remoteAddr string) {
	if isExpectedWebSocketClose(err) {
		logging.Network.Debug("websocket read closed",
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return
	}

	logging.Network.Warn("websocket read failed",
		logging.FieldError, err,
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
	)
}

func logWebSocketWriteClose(err error, roomID string, playerID string, remoteAddr string) {
	if isExpectedWebSocketClose(err) {
		logging.Network.Debug("websocket write closed",
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return
	}

	logging.Network.Error("websocket write failed", err,
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
	)
}

func isExpectedWebSocketClose(err error) bool {
	return websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
	)
}
