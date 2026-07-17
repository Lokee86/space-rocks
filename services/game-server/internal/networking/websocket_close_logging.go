package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/gorilla/websocket"
)

func logWebSocketReadClose(err error, connectionTraceID string, sessionID string, roomID string, playerID string) {
	if isExpectedWebSocketClose(err) {
		return
	}

	logging.Emit(observability.Request{
		Event: observability.EventNameGameServerReadFailed,
		Context: observability.Context{
			TraceID:   connectionTraceID,
			SessionID: sessionID,
			RoomID:    roomID,
			PlayerID:  playerID,
		},
		Fields: observability.Fields{
			"error_code":   "websocket_read_failed",
			"failure_mode": "websocket_read_failed",
		},
	})
}

func logWebSocketWriteClose(err error, connectionTraceID string, sessionID string, roomID string, playerID string) {
	if isExpectedWebSocketClose(err) {
		return
	}

	logging.Emit(observability.Request{
		Event: observability.EventNameGameServerWriteFailed,
		Context: observability.Context{
			TraceID:   connectionTraceID,
			SessionID: sessionID,
			RoomID:    roomID,
			PlayerID:  playerID,
		},
		Fields: observability.Fields{
			"error_code":   "websocket_write_failed",
			"failure_mode": "websocket_write_failed",
		},
	})
}

func isExpectedWebSocketClose(err error) bool {
	return websocket.IsCloseError(
		err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway,
		websocket.CloseNoStatusReceived,
	)
}
