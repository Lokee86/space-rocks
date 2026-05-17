package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/gorilla/websocket"
)

func main() {
	mux := http.NewServeMux()
	room := game.New()
	room.Start()

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
