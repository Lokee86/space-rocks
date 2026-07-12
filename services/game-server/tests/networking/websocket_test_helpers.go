package networkingtests

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func dialTestWebSocket(url string) (*websocket.Conn, *http.Response, error) {
	return websocket.DefaultDialer.Dial(url, http.Header{
		"Origin": []string{"https://space-rocks-client.local"},
	})
}
