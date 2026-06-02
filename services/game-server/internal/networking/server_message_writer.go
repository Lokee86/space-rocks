package networking

import (
	"github.com/gorilla/websocket"
)

func writeServerMessage(conn *websocket.Conn, message []byte, roomID string, playerID string, remoteAddr string) bool {
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		logWebSocketWriteClose(err, roomID, playerID, remoteAddr)
		return false
	}

	return true
}
