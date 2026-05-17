package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/gorilla/websocket"
)

func main() {
	mux := http.NewServeMux()
	room := game.New()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", wsHandler(room))

	fmt.Println("Server starting on :8080")

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func wsHandler(room *game.Game) http.HandlerFunc {
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

		handleConnection(conn, room)
	}
}

func handleConnection(conn *websocket.Conn, room *game.Game) {
	defer conn.Close()

	playerID := room.AddPlayer()
	defer room.RemovePlayer(playerID)

	log.Println("client connected:", playerID)
	defer log.Println("client disconnected:", playerID)

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var packet game.InputPacket
		if err := json.Unmarshal(msg, &packet); err != nil {
			log.Println("bad packet:", err)
			continue
		}

		response := room.HandlePacket(playerID, packet)
		if response == nil {
			continue
		}

		if err := conn.WriteMessage(messageType, response); err != nil {
			log.Println(err)
			break
		}
	}
}
