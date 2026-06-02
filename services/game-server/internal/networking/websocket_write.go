package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
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
			if !writeServerMessage(session.conn, message, session.currentRoomID, session.currentGamePlayerID, remoteAddr) {
				return
			}
		case <-ticker.C:
			if session.currentGamePlayerID == "" || !canSendGameplayPresentationState(session.room) {
				continue
			}

			checkRoomGameOver(session.room)

			response, ok := buildGameplayPresentationStateResponse(session, remoteAddr)
			if !ok {
				continue
			}

			if !writeServerMessage(session.conn, response, session.currentRoomID, session.currentGamePlayerID, remoteAddr) {
				return
			}
		}
	}
}
