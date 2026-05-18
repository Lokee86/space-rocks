package main

import (
	"net/http"
	"os"

	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking"
)

func main() {
	logging.Configure(os.Getenv("LOG_LEVEL"))

	mux := http.NewServeMux()
	rooms := networking.NewRoomManager()
	defer rooms.StopAll()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", networking.WebSocketHandler(rooms))

	logging.Info("server starting", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logging.Error("server stopped", err, "addr", ":8080")
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
