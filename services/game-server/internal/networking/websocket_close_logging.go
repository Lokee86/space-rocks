package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/gorilla/websocket"
)

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
