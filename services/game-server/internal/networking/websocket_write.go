package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
	"github.com/gorilla/websocket"
)

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
			if session.currentPlayerID == "" || !canSendGameplayPresentationState(session.room) {
				continue
			}

			checkRoomGameOver(session.room)

			statePacket := session.room.Game.StatePacket(session.currentPlayerID)
			response, err := packetcodec.Encode(statePacket)
			if err != nil {
				logging.Network.Error("state packet encode failed", err,
					logging.FieldRoomID, session.currentRoomID,
					logging.FieldPlayerID, session.currentPlayerID,
					logging.FieldRemoteAddr, remoteAddr,
				)
				continue
			}

			if err := session.conn.WriteMessage(websocket.TextMessage, response); err != nil {
				logWebSocketWriteClose(err, session.currentRoomID, session.currentPlayerID, remoteAddr)
				return
			}
		}
	}
}

func canSendGameplayPresentationState(room *rooms.Room) bool {
	return room != nil &&
		room.Game != nil &&
		(room.State == rooms.RoomStateInGame || room.State == rooms.RoomStateGameOver)
}

func checkRoomGameOver(room *rooms.Room) bool {
	if !room.MarkGameOverIfComplete() {
		return false
	}

	logging.Rooms.Debug("room game over detected",
		logging.FieldRoomID, room.ID,
	)
	BroadcastRoomSnapshot(room)
	return true
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
