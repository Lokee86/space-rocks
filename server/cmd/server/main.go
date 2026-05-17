package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type InputPacket struct {
	Type  string `json:"type"`
	Input struct {
		Forward bool `json:"forward"`
		Back    bool `json:"back"`
		Right   bool `json:"right"`
		Left    bool `json:"left"`
		Shoot   bool `json:"shoot"`
	} `json:"input"`
}

type Player struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func main() {
	mux := http.NewServeMux()

	player := Player{
		X: 576,
		Y: 320,
	}

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", player.wsHandler)

	fmt.Println("Server starting on :8080")

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (player *Player) wsHandler(w http.ResponseWriter, r *http.Request) {

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer conn.Close()

	log.Println("client connected")

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		var packet InputPacket

		err = json.Unmarshal(msg, &packet)
		if err != nil {
			log.Println("bad packet:", err)
			continue
		}

		returnMsg := packetHandler(packet, player)

		err = conn.WriteMessage(messageType, returnMsg)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func packetHandler(input InputPacket, player *Player) []byte {
	switch input.Type {
	case "input":
		if input.Input.Forward {
			player.Y -= 100
		}

		if input.Input.Back {
			player.Y += 100
		}

		if input.Input.Left {
			player.X -= 100
		}

		if input.Input.Right {
			player.X += 100
		}

		if input.Input.Shoot {
			log.Println("shoot")
		}
	}

	log.Printf("input: %+v player: %+v\n", input.Input, player)

	response, err := json.Marshal(player)
	if err != nil {
		log.Println(err)
		return nil
	}

	return response

}
