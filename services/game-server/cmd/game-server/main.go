package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/matchreporting"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return runWithContext(ctx)
}

func runWithContext(ctx context.Context) int {
	logging.Configure(os.Getenv(logging.EnvGlobalLevel))

	logPath, err := logging.ConfigureFileOutput("logs/game-server", "game-server")
	if err != nil {
		logging.Server.Warn("server structured log file unavailable", logging.FieldError, err)
	} else {
		logging.Server.Info("server structured log file configured", "path", logPath)
	}

	defer func() {
		if err := logging.CloseFileOutput(); err != nil {
			logging.Server.Error("server structured log file close failed", err)
		}
	}()

	mux := http.NewServeMux()
	diagnosticService, err := registerDiagnosticAggregator(mux)
	if err != nil {
		logging.Server.Error("diagnostic aggregator initialization failed", err)
		return 1
	}
	defer closeDiagnosticAggregator(diagnosticService)
	rooms := networking.NewRoomManager()
	defer rooms.StopAll()

	webRTCTransportConfig := buildWebRTCTransportConfigFromEnv()
	networking.SetWebRTCTransportConfig(webRTCTransportConfig)
	logging.Server.Info(
		"web rtc transport config loaded",
		"advertised_ip_count", len(webRTCTransportConfig.AdvertisedIPs),
		"udp_port_range_configured", webRTCTransportConfig.UDPPortMin != 0 && webRTCTransportConfig.UDPPortMax != 0,
	)

	playerDataRuntime, err := buildPlayerDataRuntime()
	if err != nil {
		logging.Server.Error("player-data runtime initialization failed", err)
		return 1
	}
	playerDataSink := newPlayerDataSink(playerDataRuntime)

	reporter, err := matchreporting.NewRuntimeReporter(playerDataSink)
	if err != nil {
		logging.Server.Error("player-data reporter initialization failed", err)
		return 1
	}
	authVerifier := buildAuthVerifierFromEnv()

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", networking.WebSocketHandlerWithAuthAndReporter(rooms, authVerifier, reporter))

	playerDataProfileHandler := newPlayerDataProfileHTTPHandler(playerDataRuntime, authVerifier)
	playerDataLocalProfilesHandler := newPlayerDataLocalProfilesHTTPHandler(playerDataRuntime)
	mux.Handle("POST /api/player-data/profile", playerDataProfileHandler)
	mux.Handle("GET /api/player-data/local-profiles", playerDataLocalProfilesHandler)
	mux.Handle("POST /api/player-data/local-profiles", playerDataLocalProfilesHandler)
	mux.Handle("PUT /api/player-data/local-profiles/{local_profile_id}", playerDataLocalProfilesHandler)
	mux.Handle("DELETE /api/player-data/local-profiles/{local_profile_id}", playerDataLocalProfilesHandler)
	mux.Handle("GET /api/player-data/local-profiles/default", playerDataLocalProfilesHandler)
	mux.Handle("PUT /api/player-data/local-profiles/default", playerDataLocalProfilesHandler)

	logging.Server.Info("server starting", "addr", ":8080")
	server := newHTTPServer(mux)
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		logging.Server.Error("server stopped", err, "addr", ":8080")
		return 1
	}
	if err := serveHTTPServer(ctx, server, listener, 5*time.Second); err != nil {
		logging.Server.Error("server stopped", err, "addr", ":8080")
		return 1
	}
	return 0
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
