package networking

import (
	"net/http"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
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
	defer session.clearWebRTCTransport()
	defer session.conn.Close()
	defer session.leaveDisconnectedRoom()

	context := session.sessionContext()
	logging.Network.Debug("websocket connected",
		logging.FieldRoomID, context.RoomID,
		logging.FieldPlayerID, context.GamePlayerID,
		"session_id", session.sessionID,
		"current_room_id", context.RoomID,
		logging.FieldRemoteAddr, remoteAddr,
	)
	defer func() {
		context := session.sessionContext()
		logging.Network.Debug("websocket disconnected",
			logging.FieldRoomID, context.RoomID,
			logging.FieldPlayerID, context.GamePlayerID,
			"session_id", session.sessionID,
			logging.FieldRemoteAddr, remoteAddr,
		)
	}()

	readErr := make(chan error, 1)
	gameplayLifecycleDone := make(chan struct{})
	defer close(gameplayLifecycleDone)

	go readClientInput(session, remoteAddr, readErr)
	go tickSessionGameplayLifecycle(session, gameplayLifecycleDone)

	writeServerMessages(session, remoteAddr, readErr)
}
