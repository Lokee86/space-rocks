package networking

import (
	"strings"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/outbound"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/packetmetrics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
)

const debugStatusWriteIntervalTicks = 8

var canSendDebugShapeCatalog = outbound.CanSendDebugShapeCatalog
var buildDebugShapeCatalogResponse = outbound.BuildDebugShapeCatalogResponse

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
			if !writeGameplayLaneProtocolMessage(session, remoteAddr) {
				return
			}
		}
	}
}

func writeGameplayLaneProtocolMessage(session *webSocketSession, remoteAddr string) bool {
	if session.room == nil || session.currentGamePlayerID == "" || session.room.GameInstance() == nil {
		return true
	}

	if !maybeWriteDebugShapeCatalog(session, remoteAddr) {
		return false
	}

	if session.realtimeState.ReceiverID == "" || session.realtimeState.ReceiverID != session.currentGamePlayerID {
		session.realtimeState = realtime.NewRealtimeSessionState(session.currentGamePlayerID)
	}

	result, err := realtime.BuildActiveRealtimeResultForGame(session.room.GameInstance(), session.currentGamePlayerID, session.realtimeState)
	if err != nil {
		logging.Network.Error("lane protocol gameplay build failed", err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return false
	}

	drainedEventCount := 0
	for _, encoded := range result.EncodedLanePackets {
		candidate := encoded.Candidate
		encodedPacket := encoded.Encoded
		if len(encodedPacket) == 0 {
			continue
		}
		if candidate.Lane() == realtime.LaneControl {
			logging.Network.Warn("lane protocol gameplay webrtc control lane is websocket-owned",
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				logging.FieldRemoteAddr, remoteAddr,
				"lane", candidate.Lane(),
				"transport", "webrtc",
			)
			return false
		}
		if session.webrtcTransport == nil {
			logging.Network.Warn("lane protocol gameplay webrtc transport missing",
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				logging.FieldRemoteAddr, remoteAddr,
				"lane", candidate.Lane(),
				"transport", "webrtc",
			)
			continue
		}
		if !session.webrtcTransport.Ready() {
			logging.Network.Warn("lane protocol gameplay webrtc transport not ready",
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				logging.FieldRemoteAddr, remoteAddr,
				"lane", candidate.Lane(),
				"transport", "webrtc",
			)
			continue
		}
		channelLabel, ok := webRTCGameplayChannelLabelForLane(string(candidate.Lane()))
		if !ok {
			logging.Network.Warn("lane protocol gameplay webrtc lane channel missing",
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				logging.FieldRemoteAddr, remoteAddr,
				"lane", candidate.Lane(),
				"transport", "webrtc",
			)
			return false
		}
		if err := session.webrtcTransport.SendEncodedLaneJSON(string(candidate.Lane()), encodedPacket); err != nil {
			logging.Network.Error("lane protocol gameplay webrtc write failed", err,
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				logging.FieldRemoteAddr, remoteAddr,
				"lane", candidate.Lane(),
				"transport", "webrtc",
				"channel", channelLabel,
			)
			return false
		}
		diagnostics := realtime.CandidateWriteDiagnosticsFor(candidate, session.realtimeState, len(encodedPacket))
		logging.Network.Debug("lane protocol gameplay wire packet written",
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			logging.FieldRemoteAddr, remoteAddr,
			"transport", "webrtc",
			"channel", channelLabel,
			"packet_family", diagnostics.PacketFamily,
			"candidate_lane", diagnostics.Lane,
			"candidate_kind", diagnostics.Kind,
			"sequence", diagnostics.Sequence,
			"baseline_id", diagnostics.BaselineID,
			"snapshot_id", diagnostics.SnapshotID,
			"snapshot_kind", diagnostics.SnapshotKind,
			"chunk_index", diagnostics.ChunkIndex,
			"chunk_count", diagnostics.ChunkCount,
			"is_final_chunk", diagnostics.IsFinalChunk,
			"encoded_bytes", len(encodedPacket),
		)
		if candidate.Kind() == realtime.RealtimeLaneCandidateKindEventBatch {
			if drained := drainActiveEventBatchAfterWrite(session.room.GameInstance(), session.currentGamePlayerID, result.EventBatchEventIDs); len(drained) > 0 {
				drainedEventCount += len(drained)
			}
		}

		if metadata, ok := candidate.Metadata(); ok {
			persistedMetadata := realtime.AdvanceMetadataForSuccessfulWrite(candidate.Lane(), metadata)
			session.realtimeState.UpdateLane(candidate.Lane(), persistedMetadata)
			if projection, ok := realtime.CandidateProjection(candidate); ok {
				session.realtimeState.StoreBaselineProjection(candidate.Lane(), projection)
			}
			if metadata.IsFinalChunk && candidate.Kind() == realtime.RealtimeLaneCandidateKindFull {
				session.realtimeState.MarkBaselineReady(candidate.Lane())
			}
		}
	}

	if len(result.MetricSummaries) == 0 && result.TotalEncodedBytes == 0 {
		return true
	}

	logging.Network.Debug("lane protocol gameplay written",
		logging.FieldRoomID, session.currentRoomID,
		logging.FieldPlayerID, session.currentGamePlayerID,
		logging.FieldRemoteAddr, remoteAddr,
		"lane_packet_families", lanePacketFamilySummary(result.MetricSummaries),
		"baseline_full_count", countLaneCandidateKinds(result.SelectedCandidates, realtime.RealtimeLaneCandidateKindFull),
		"event_batch_written", len(result.EventBatchEventIDs) > 0,
		"event_batch_drained_count", drainedEventCount,
		"packet_count", len(result.MetricSummaries),
		"encoded_bytes", result.TotalEncodedBytes,
	)
	return true
}

func countLaneCandidateKinds(candidates []realtime.RealtimeLaneCandidate, kind realtime.RealtimeLaneCandidateKind) int {
	count := 0
	for _, candidate := range candidates {
		if candidate.Kind() == kind {
			count++
		}
	}
	return count
}

func lanePacketFamilySummary(records []packetmetrics.PacketMetricRecord) string {
	if len(records) == 0 {
		return ""
	}

	families := make([]string, 0, len(records))
	for _, record := range records {
		families = append(families, record.PacketFamily)
	}
	return strings.Join(families, ",")
}

func drainActiveEventBatchAfterWrite(gameInstance *game.Game, playerID string, eventIDs []string) []game.PendingPresentationEvent {
	if gameInstance == nil || len(eventIDs) == 0 {
		return nil
	}

	return gameInstance.DrainPendingPresentationEvents(playerID, eventIDs...)
}

func maybeWriteDebugShapeCatalog(session *webSocketSession, remoteAddr string) bool {
	if session == nil || session.room == nil {
		return true
	}
	if session.debugShapeCatalogSentRoomID == session.currentRoomID {
		return true
	}
	if !canSendDebugShapeCatalog(session.room) {
		return true
	}

	response, ok := buildDebugShapeCatalogResponse(session.room, session.currentRoomID, remoteAddr)
	if !ok {
		return true
	}
	if !outbound.WriteServerMessage(session.conn, response, func(err error) {
		logWebSocketWriteClose(err, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
	}) {
		return false
	}

	logging.Network.Debug("debug shape catalog written",
		logging.FieldRoomID, session.currentRoomID,
		logging.FieldPlayerID, session.currentGamePlayerID,
		logging.FieldRemoteAddr, remoteAddr,
	)
	session.debugShapeCatalogSentRoomID = session.currentRoomID
	return true
}

