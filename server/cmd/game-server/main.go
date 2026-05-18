package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Lokee86/space-rocks/server/internal/networking"
)

func main() {
	mux := http.NewServeMux()
	rooms := networking.NewRoomManager()
	defer rooms.StopAll()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", networking.WebSocketHandler(rooms))

	fmt.Println("Server starting on :8080")

	log.Fatal(http.ListenAndServe(":8080", mux))

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
