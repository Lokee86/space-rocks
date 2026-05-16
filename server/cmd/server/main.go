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

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", wsHandler)

	fmt.Println("Server starting on :8080")

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
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

		log.Printf("Decoded packet: %+v\n", packet)

		err = conn.WriteMessage(messageType, msg)
		if err != nil {
			log.Println(err)
			break
		}
	}
}
