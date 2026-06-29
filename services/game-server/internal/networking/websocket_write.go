package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking/outbound"
	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
	"github.com/Lokee86/space-rocks/server/internal/protocol/realtime"
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
	lastDebugShapeCatalogRoomID := ""

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

			response, packetMetrics, ok := outbound.BuildGameplayPresentationStateResponse(session.room, session.currentGamePlayerID, session.currentRoomID, remoteAddr)
			if !ok {
				continue
			}

			writeStarted := time.Now()
			if !outbound.WriteServerMessage(session.conn, response, func(err error) {
				logWebSocketWriteClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
			}) {
				return
			}
			packetmetrics.LogSlowGameplayPresentationWrite(time.Since(writeStarted), packetMetrics, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
			if devtools.Enabled() {
				runShadowRealtimeMeasurement(session, remoteAddr)
			}
			lastDebugShapeCatalogRoomID = writeDebugShapeCatalogMessage(session, remoteAddr, lastDebugShapeCatalogRoomID)
			debugStatusTick++
			if debugStatusTick >= debugStatusWriteIntervalTicks {
				debugStatusTick = 0
				writeDebugStatusMessage(session, remoteAddr)
			}
		}
	}
}

func runShadowRealtimeMeasurement(session *webSocketSession, remoteAddr string) {
	if session.room == nil || session.currentGamePlayerID == "" || session.room.GameInstance() == nil {
		return
	}

	snapshot := session.room.GameInstance().GameplayPresentationSnapshot(session.currentGamePlayerID)
	result := realtime.BuildShadowRealtimeResult(snapshot, realtime.NewRealtimeSessionState(session.currentGamePlayerID))
	packetmetrics.LogShadowLaneMetrics(realtime.ShadowLaneMetricRecords(result), session.currentRoomID, session.currentGamePlayerID, remoteAddr)
	fields := []any{
		logging.FieldRoomID, session.currentRoomID,
		logging.FieldPlayerID, session.currentGamePlayerID,
		logging.FieldRemoteAddr, remoteAddr,
	}
	fields = append(fields, realtime.ShadowRealtimeSummaryFields(result)...)
	logging.Network.Debug("shadow realtime summary", fields...)
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

func writeDebugShapeCatalogMessage(session *webSocketSession, remoteAddr string, lastSentRoomID string) string {
	if session.currentRoomID == "" || session.currentRoomID == lastSentRoomID || !outbound.CanSendDebugShapeCatalog(session.room) {
		return lastSentRoomID
	}

	response, ok := outbound.BuildDebugShapeCatalogResponse(session.room, session.currentRoomID, remoteAddr)
	if !ok {
		return lastSentRoomID
	}

	if !outbound.WriteServerMessage(session.conn, response, func(err error) {
		logWebSocketWriteClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
	}) {
		return lastSentRoomID
	}

	return session.currentRoomID
}
