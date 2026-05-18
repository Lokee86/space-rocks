package networking

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game"
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
			log.Println(err)
			return
		}

		room, leaveRoom := rooms.Join(r.URL.Query().Get(RoomIDQueryParam))
		defer leaveRoom()
		handleConnection(conn, room.Game)
	}
}

func handleConnection(conn *websocket.Conn, room *game.Game) {
	defer conn.Close()

	playerID := room.AddPlayer()
	defer room.RemovePlayer(playerID)

	log.Println("client connected:", playerID)
	defer log.Println("client disconnected:", playerID)

	readErr := make(chan error, 1)
	go readClientInput(conn, room, playerID, readErr)

	writeServerState(conn, room, playerID, readErr)
}

func readClientInput(conn *websocket.Conn, room *game.Game, playerID string, readErr chan<- error) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			readErr <- err
			return
		}

		var packet game.ClientPacket
		if err := json.Unmarshal(msg, &packet); err != nil {
			log.Println("bad packet:", err)
			continue
		}

		room.HandlePacket(playerID, packet)
	}
}

func writeServerState(conn *websocket.Conn, room *game.Game, playerID string, readErr <-chan error) {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	for {
		select {
		case err := <-readErr:
			log.Println(err)
			return
		case <-ticker.C:
			response := room.State(playerID)
			if response == nil {
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, response); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
