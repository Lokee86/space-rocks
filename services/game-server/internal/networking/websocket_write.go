package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/networking/outbound"
)

const debugStatusWriteIntervalTicks = 8

func writeServerMessages(
	session *webSocketSession,
	remoteAddr string,
	readErr <-chan error,
) {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	debugStatusTick := 0

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

			response, ok := outbound.BuildGameplayPresentationStateResponse(session.room, session.currentGamePlayerID, session.currentRoomID, remoteAddr)
			if !ok {
				continue
			}

			writeStarted := time.Now()
			if !outbound.WriteServerMessage(session.conn, response, func(err error) {
				logWebSocketWriteClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
			}) {
				return
			}
			outbound.LogSlowGameplayPresentationWrite(time.Since(writeStarted), session.currentRoomID, session.currentGamePlayerID, remoteAddr)
			debugStatusTick++
			if debugStatusTick >= debugStatusWriteIntervalTicks {
				debugStatusTick = 0
				writeDebugStatusMessage(session, remoteAddr)
			}
		}
	}
}

func writeDebugStatusMessage(session *webSocketSession, remoteAddr string) {
	if session.currentGamePlayerID == "" || !outbound.CanSendDebugStatus(session.room) {
		return
	}

	response, ok := outbound.BuildDebugStatusResponse(session.room, session.currentGamePlayerID, session.currentRoomID, remoteAddr)
	if !ok {
		return
	}

	if !outbound.WriteServerMessage(session.conn, response, func(err error) {
		logWebSocketWriteClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
	}) {
		return
	}
}
