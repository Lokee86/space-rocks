package outbound

import (
	"time"

	"github.com/gorilla/websocket"
)

const serverMessageWriteTimeout = 10 * time.Second

type serverMessageWriter interface {
	SetWriteDeadline(time.Time) error
	WriteMessage(int, []byte) error
}

func writeServerMessage(conn serverMessageWriter, message []byte, onWriteClose func(error)) bool {
	if err := conn.SetWriteDeadline(time.Now().Add(serverMessageWriteTimeout)); err != nil {
		if onWriteClose != nil {
			onWriteClose(err)
		}
		return false
	}
	if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
		if onWriteClose != nil {
			onWriteClose(err)
		}
		return false
	}
	return true
}

func WriteServerMessage(
	conn *websocket.Conn,
	message []byte,
	onWriteClose func(error),
) bool {
	return writeServerMessage(conn, message, onWriteClose)
}
