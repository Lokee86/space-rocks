package outbound

import "github.com/gorilla/websocket"

func WriteServerMessage(
	conn *websocket.Conn,
	message []byte,
	onWriteClose func(error),
) bool {
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		if onWriteClose != nil {
			onWriteClose(err)
		}
		return false
	}

	return true
}
