package main

import (
	"net/http"
	"os"

	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking"
)

func main() {
	logging.Configure(os.Getenv(logging.EnvGlobalLevel))

	mux := http.NewServeMux()
	rooms := networking.NewRoomManager()
	defer rooms.StopAll()
	playerDataRuntime, err := buildPlayerDataRuntime()
	if err != nil {
		logging.Server.Error("player-data runtime initialization failed", err)
		os.Exit(1)
	}
	_ = playerDataRuntime // wiring happens in a later phase
	authVerifier := buildAuthVerifierFromEnv()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", networking.WebSocketHandlerWithAuth(rooms, authVerifier))

	logging.Server.Info("server starting", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logging.Server.Error("server stopped", err, "addr", ":8080")
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
