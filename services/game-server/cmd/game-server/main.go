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
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/matchreporting"
	"github.com/Lokee86/space-rocks/services/game-server/internal/measurement"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking"
	"github.com/Lokee86/space-rocks/services/game-server/internal/playerinventory"
	servertooling "github.com/Lokee86/space-rocks/services/game-server/internal/tooling"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
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
	startupTraceID := uuid.NewString()
	shutdownTraceID := uuid.NewString()
	normalShutdown := false
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
	logging.Emit(observability.Request{
		Event:   observability.EventNameServiceStarting,
		Context: observability.Context{TraceID: startupTraceID},
		Fields:  observability.Fields{"reason_code": "process_start"},
	})
	_, err = logging.ConfigureFileOutput(runtimePath("logs/game-server"), "game-server")
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameObservabilityUnavailable,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"failure_mode": "logging_file_open_failed"},
		})
	} else if status := logging.Status(); status.Degraded {
		logging.Emit(observability.Request{
			Event:   observability.EventNameObservabilityUnavailable,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"failure_mode": "logging_runtime_degraded"},
		})
	}
	defer func() {
		if err := logging.CloseFileOutput(); err != nil {
			logging.Emit(observability.Request{
				Event:   observability.EventNameObservabilityUnavailable,
				Context: observability.Context{TraceID: shutdownTraceID},
				Fields:  observability.Fields{"failure_mode": "logging_file_close_failed"},
			})
		}
	}()
	defer func() {
		if normalShutdown {
			logging.Emit(observability.Request{
				Event:   observability.EventNameServiceStopped,
				Context: observability.Context{TraceID: shutdownTraceID},
			})
		}
	}()

	playerlogging.Configure(os.Getenv(playerlogging.EnvGlobalLevel))
	if err := playerlogging.ConfigureRuntime(playerDataIdentity); err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameDependencyInitializationFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"dependency": "player_data_logging", "failure_mode": "configuration_failed"},
		})
		return 1
	}
	_, err = playerlogging.ConfigureFileOutput(runtimePath("logs/player-data"), "player-data")
	if err != nil {
		emitPlayerDataObservabilityUnavailable("open_failed")
	} else if status := playerlogging.Status(); status.Degraded {
		emitPlayerDataObservabilityUnavailable("runtime_degraded")
	}
	defer func() {
		if err := playerlogging.CloseFileOutput(); err != nil {
			logging.Emit(observability.Request{
				Event:   observability.EventNameObservabilityUnavailable,
				Context: observability.Context{TraceID: shutdownTraceID},
				Fields:  observability.Fields{"failure_mode": "player_data_logging_close_failed"},
			})
		}
	}()

	mux := http.NewServeMux()
	registerRuntimeScenarioPprof(mux)
	diagnosticService, err := registerDiagnosticAggregator(mux)
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameDependencyInitializationFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"dependency": "diagnostic_aggregator", "failure_mode": "initialization_failed"},
		})
		return 1
	}
	defer closeDiagnosticAggregator(diagnosticService, shutdownTraceID)
	gameFactory, err := runtimeScenarioGameFactoryFromEnv()
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameConfigurationInvalid,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"configuration_key": runtimeScenarioSeedEnv, "failure_mode": "invalid_value"},
		})
		return 1
	}
	rooms := networking.NewRoomManagerWithGameFactory(gameFactory)
	defer rooms.StopAll()
	measurementController := servertooling.NewController(servertooling.Dependencies{
		Rooms:        rooms,
		BuildVersion: gameIdentity.Version,
		ReportWriter: measurement.NewReportWriter(runtimeScenarioMeasurementPath(runtimePath("measurement-results/game-server"))),
	})

	webRTCTransportConfig := buildWebRTCTransportConfigFromEnv()
	networking.SetWebRTCTransportConfig(webRTCTransportConfig)

	playerDataRuntime, err := buildPlayerDataRuntime()
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameDependencyInitializationFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"dependency": "player_data_runtime", "failure_mode": "initialization_failed"},
		})
		return 1
	}
	playerDataSink := newPlayerDataSink(playerDataRuntime)
	inventoryClient, err := playerinventory.NewRuntimeClient(playerDataSink)
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameDependencyInitializationFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"dependency": "player_inventory_client", "failure_mode": "initialization_failed"},
		})
		return 1
	}
	buildService, err := playerbuild.NewService(inventoryClient, playerbuild.DefaultCatalog())
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameDependencyInitializationFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"dependency": "player_build_service", "failure_mode": "initialization_failed"},
		})
		return 1
	}

	reporter, err := matchreporting.NewRuntimeReporter(playerDataSink)
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameDependencyInitializationFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"dependency": "player_data_reporter", "failure_mode": "initialization_failed"},
		})
		return 1
	}
	authVerifier := buildAuthVerifierFromEnv(startupTraceID)

	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ws", networking.WebSocketHandlerWithAuthReporterToolingAndBuilds(rooms, authVerifier, reporter, measurementController, measurementController, buildService))

	playerDataProfileHandler := newPlayerDataProfileHTTPHandler(playerDataRuntime, authVerifier)
	playerDataLocalProfilesHandler := newPlayerDataLocalProfilesHTTPHandler(playerDataRuntime)
	mux.Handle("POST /api/player-data/profile", playerDataProfileHandler)
	mux.Handle("GET /api/player-data/local-profiles", playerDataLocalProfilesHandler)
	mux.Handle("POST /api/player-data/local-profiles", playerDataLocalProfilesHandler)
	mux.Handle("PUT /api/player-data/local-profiles/{local_profile_id}", playerDataLocalProfilesHandler)
	mux.Handle("DELETE /api/player-data/local-profiles/{local_profile_id}", playerDataLocalProfilesHandler)
	mux.Handle("GET /api/player-data/local-profiles/default", playerDataLocalProfilesHandler)
	mux.Handle("PUT /api/player-data/local-profiles/default", playerDataLocalProfilesHandler)

	server := newHTTPServer(httpapi.WithRequestContext(mux))
	listener, err := net.Listen("tcp", serverListenAddress())
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameServiceRuntimeFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"failure_mode": "listen_failed"},
		})
		return 1
	}
	logging.Emit(observability.Request{
		Event:   observability.EventNameServiceStarted,
		Context: observability.Context{TraceID: startupTraceID},
	})
	if err := serveHTTPServer(ctx, server, listener, 5*time.Second, func() {
		logging.Emit(observability.Request{
			Event:   observability.EventNameServiceStopping,
			Context: observability.Context{TraceID: shutdownTraceID},
		})
	}); err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameServiceRuntimeFailed,
			Context: observability.Context{TraceID: startupTraceID},
			Fields:  observability.Fields{"failure_mode": "serve_failed"},
		})
		return 1
	}
	normalShutdown = true
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
