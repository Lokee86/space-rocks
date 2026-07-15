package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Lokee86/space-rocks/player-data/httpapi"
	playerlogging "github.com/Lokee86/space-rocks/player-data/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/matchreporting"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
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
	gameIdentity, err := loadLoggingIdentity(logging.ServiceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "game-server logging configuration failed: %v\n", err)
		return 1
	}
	playerDataIdentity, err := loadLoggingIdentity(playerlogging.ServiceName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "player-data logging configuration failed: %v\n", err)
		return 1
	}

	logging.Configure(os.Getenv(logging.EnvGlobalLevel))
	if err := logging.ConfigureRuntime(gameIdentity); err != nil {
		fmt.Fprintf(os.Stderr, "game-server logging configuration failed: %v\n", err)
		return 1
	}
	logPath, err := logging.ConfigureFileOutput("logs/game-server", "game-server")
	if err != nil {
		logging.Server.Warn("server structured log file unavailable", logging.FieldError, err)
	} else if status := logging.Status(); status.Degraded {
		logging.Server.Warn("server structured log file unavailable", logging.FieldError, status.LastError)
	} else {
		logging.Server.Info("server structured log file configured", "path", logPath)
	}
	defer func() {
		if err := logging.CloseFileOutput(); err != nil {
			logging.Server.Error("server structured log file close failed", err)
		}
	}()

	playerlogging.Configure(os.Getenv(playerlogging.EnvGlobalLevel))
	if err := playerlogging.ConfigureRuntime(playerDataIdentity); err != nil {
		logging.Server.Error("player-data logging configuration failed", err)
		return 1
	}
	_, err = playerlogging.ConfigureFileOutput("logs/player-data", "player-data")
	if err != nil {
		emitPlayerDataObservabilityUnavailable("open_failed")
	} else if status := playerlogging.Status(); status.Degraded {
		emitPlayerDataObservabilityUnavailable("runtime_degraded")
	}
	defer func() {
		if err := playerlogging.CloseFileOutput(); err != nil {
			logging.Server.Error("player-data structured log file close failed", err)
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
	server := newHTTPServer(httpapi.WithRequestContext(mux))
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

func emitPlayerDataObservabilityUnavailable(failureMode string) observability.Result {
	return playerlogging.Emit(observability.Request{
		Event:   observability.EventNameObservabilityUnavailable,
		Message: "player-data structured logging unavailable",
		Fields:  observability.Fields{"failure_mode": failureMode},
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}
