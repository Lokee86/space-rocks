package networking

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/gorilla/websocket"
)

const RoomIDQueryParam = "room_id"

func WebSocketHandler(rooms *RoomManager) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Error("websocket upgrade failed", err, logging.FieldRemoteAddr, r.RemoteAddr)
			return
		}

		room, leaveRoom := rooms.Join(r.URL.Query().Get(RoomIDQueryParam))
		defer leaveRoom()
		handleConnection(conn, room, r.RemoteAddr)
	}
}

func handleConnection(conn *websocket.Conn, room *Room, remoteAddr string) {
	defer conn.Close()

	playerID := room.Game.AddPlayer()
	defer room.Game.RemovePlayer(playerID)

	logging.Info("websocket connected",
		logging.FieldRoomID, room.ID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
	)
	defer logging.Info("websocket disconnected",
		logging.FieldRoomID, room.ID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
	)

	readErr := make(chan error, 1)
	go readClientInput(conn, room.Game, playerID, room.ID, remoteAddr, readErr)

	writeServerState(conn, room.Game, playerID, room.ID, remoteAddr, readErr)
}

func readClientInput(
	conn *websocket.Conn,
	room *game.Game,
	playerID string,
	roomID string,
	remoteAddr string,
	readErr chan<- error,
) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			readErr <- err
			return
		}

		var packet game.ClientPacket
		if err := json.Unmarshal(msg, &packet); err != nil {
			logging.Warn("websocket packet decode failed",
				logging.FieldError, err,
				logging.FieldRoomID, roomID,
				logging.FieldPlayerID, playerID,
				logging.FieldRemoteAddr, remoteAddr,
			)
			continue
		}

		room.HandlePacket(playerID, packet)
	}
}

func writeServerState(
	conn *websocket.Conn,
	room *game.Game,
	playerID string,
	roomID string,
	remoteAddr string,
	readErr <-chan error,
) {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	for {
		select {
		case err := <-readErr:
			logWebSocketReadClose(err, roomID, playerID, remoteAddr)
			return
		case <-ticker.C:
			response := room.State(playerID)
			if response == nil {
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
				logWebSocketWriteClose(err, roomID, playerID, remoteAddr)
				return
			}
		}
	}
}

func logWebSocketReadClose(err error, roomID string, playerID string, remoteAddr string) {
	if isExpectedWebSocketClose(err) {
		logging.Info("websocket read closed",
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return
	}

	logging.Warn("websocket read failed",
		logging.FieldError, err,
		logging.FieldRoomID, roomID,
		logging.FieldPlayerID, playerID,
		logging.FieldRemoteAddr, remoteAddr,
	)
}

func logWebSocketWriteClose(err error, roomID string, playerID string, remoteAddr string) {
	if isExpectedWebSocketClose(err) {
		logging.Info("websocket write closed",
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return
	}

	logging.Error("websocket write failed", err,
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
