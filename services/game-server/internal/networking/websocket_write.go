package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/networking/outbound"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func writeServerMessages(
	session *webSocketSession,
	remoteAddr string,
	readErr <-chan error,
) {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	for {
		select {
		case err := <-readErr:
			logWebSocketReadClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
			return
		case message := <-session.outbound:
			if !outbound.WriteServerMessage(session.conn, message, func(err error) {
				logWebSocketWriteClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
			}) {
				return
			}
		case <-ticker.C:
			if session.currentGamePlayerID == "" || !outbound.CanSendGameplayPresentationState(session.room) {
				continue
			}

			// Lifecycle tick: advances game-over state before building the presentation payload.
			rooms.TickRoomGameOverLifecycle(session.room, BroadcastRoomSnapshot)

			response, ok := outbound.BuildGameplayPresentationStateResponse(session.room, session.currentGamePlayerID, session.currentRoomID, remoteAddr)
			if !ok {
				continue
			}

			if !outbound.WriteServerMessage(session.conn, response, func(err error) {
				logWebSocketWriteClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
			}) {
				return
			}
		}
	}
}
