package networking

import (
	"net/http"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/playerbuild"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	toolingrouter "github.com/Lokee86/space-rocks/services/game-server/internal/networking/tooling"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const webSocketReadLimit = 256 * 1024

func WebSocketHandler(roomManager *rooms.RoomManager) http.HandlerFunc {
	return WebSocketHandlerWithAuth(roomManager, nil)
}

func WebSocketHandlerWithAuth(roomManager *rooms.RoomManager, verifier TokenVerifier) http.HandlerFunc {
	return WebSocketHandlerWithAuthAndReporter(roomManager, verifier, rooms.NoopMatchResultReporter{})
}

func WebSocketHandlerWithAuthAndReporter(roomManager *rooms.RoomManager, verifier TokenVerifier, reporter rooms.MatchResultReporter) http.HandlerFunc {
	return WebSocketHandlerWithAuthReporterToolingAndBuilds(roomManager, verifier, reporter, nil, nil, nil)
}

func WebSocketHandlerWithAuthAndReporterAndTooling(roomManager *rooms.RoomManager, verifier TokenVerifier, reporter rooms.MatchResultReporter, measurementController toolingrouter.MeasurementController, telemetryProvider toolingrouter.TelemetryProvider) http.HandlerFunc {
	return WebSocketHandlerWithAuthReporterToolingAndBuilds(roomManager, verifier, reporter, measurementController, telemetryProvider, nil)
}

func WebSocketHandlerWithAuthReporterAndBuilds(roomManager *rooms.RoomManager, verifier TokenVerifier, reporter rooms.MatchResultReporter, buildService *playerbuild.Service) http.HandlerFunc {
	return WebSocketHandlerWithAuthReporterToolingAndBuilds(roomManager, verifier, reporter, nil, nil, buildService)
}

func WebSocketHandlerWithAuthReporterToolingAndBuilds(roomManager *rooms.RoomManager, verifier TokenVerifier, reporter rooms.MatchResultReporter, measurementController toolingrouter.MeasurementController, telemetryProvider toolingrouter.TelemetryProvider, buildService *playerbuild.Service) http.HandlerFunc {
	originPolicy := newWebSocketOriginPolicy()
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return originPolicy.allows(r)
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		upgradeTraceID := uuid.NewString()
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			logging.Emit(observability.Request{
				Event:   observability.EventNameGameServerConnectionUpgradeFailed,
				Context: observability.Context{TraceID: upgradeTraceID},
				Fields: observability.Fields{
					"error_code":   "websocket_upgrade_failed",
					"failure_mode": "websocket_upgrade_failed",
				},
			})
			return
		}
		conn.SetReadLimit(webSocketReadLimit)

		session := newWebSocketSessionWithToolingAndBuilds(conn, roomManager, verifier, reporter, measurementController, telemetryProvider, buildService)
		handleConnection(session, r.RemoteAddr)
	}
}

func handleConnection(session *webSocketSession, remoteAddr string) {
	defer session.closeTooling()
	defer session.clearWebRTCTransport()
	defer session.conn.Close()
	defer session.leaveDisconnectedRoom()

	context := session.sessionContext()
	logging.Emit(observability.Request{
		Event: observability.EventNameGameServerClientConnected,
		Context: observability.Context{
			TraceID:   session.connectionTraceID,
			SessionID: session.sessionID,
			RoomID:    context.RoomID,
			PlayerID:  context.GamePlayerID,
		},
		Fields: observability.Fields{"reason_code": "connection_established"},
	})
	defer func() {
		context := session.sessionContext()
		logging.Emit(observability.Request{
			Event: observability.EventNameGameServerClientDisconnected,
			Context: observability.Context{
				TraceID:   session.connectionTraceID,
				SessionID: session.sessionID,
				RoomID:    context.RoomID,
				PlayerID:  context.GamePlayerID,
			},
			Fields: observability.Fields{"reason_code": "connection_closed"},
		})
	}()

	readErr := make(chan error, 1)
	gameplayLifecycleDone := make(chan struct{})
	defer close(gameplayLifecycleDone)

	go readClientInput(session, remoteAddr, readErr)
	go tickSessionGameplayLifecycle(session, gameplayLifecycleDone)

	writeServerMessages(session, remoteAddr, readErr)
}
