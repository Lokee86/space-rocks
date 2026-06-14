package main

import (
	"net/http"
	"os"

	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/matchreporting"
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
	playerDataSink := newPlayerDataSink(playerDataRuntime)
	reporter, err := matchreporting.NewRuntimeReporter(playerDataSink)
	if err != nil {
		logging.Server.Error("player-data reporter initialization failed", err)
		os.Exit(1)
	}
	authVerifier := buildAuthVerifierFromEnv()

	// Core server routes.
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", networking.WebSocketHandlerWithAuthAndReporter(rooms, authVerifier, reporter))

	// Player-data routes.
	playerDataProfileHandler := newPlayerDataProfileHTTPHandler(playerDataRuntime, authVerifier)
	playerDataLocalProfilesHandler := newPlayerDataLocalProfilesHTTPHandler(playerDataRuntime)
	mux.Handle("POST /api/player-data/profile", playerDataProfileHandler)
	mux.Handle("GET /api/player-data/local-profiles", playerDataLocalProfilesHandler)
	mux.Handle("POST /api/player-data/local-profiles", playerDataLocalProfilesHandler)
	mux.Handle("DELETE /api/player-data/local-profiles/{local_profile_id}", playerDataLocalProfilesHandler)
	mux.Handle("GET /api/player-data/local-profiles/default", playerDataLocalProfilesHandler)
	mux.Handle("PUT /api/player-data/local-profiles/default", playerDataLocalProfilesHandler)

	logging.Server.Info("server starting", "addr", ":8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logging.Server.Error("server stopped", err, "addr", ":8080")
		os.Exit(1)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
