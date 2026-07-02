package networking

import (
	"fmt"
	"strings"
	"time"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking/outbound"
	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
	"github.com/Lokee86/space-rocks/server/internal/protocol/realtime"
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
	for _, candidate := range result.SelectedCandidates {
		encodedPacket := result.EncodedPackets[candidate.Lane]
		if len(encodedPacket) == 0 {
			continue
		}
		if !outbound.WriteServerMessage(session.conn, encodedPacket, func(writeErr error) {
			logging.Network.Error("lane protocol gameplay write failed", writeErr,
				logging.FieldRoomID, session.currentRoomID,
				logging.FieldPlayerID, session.currentGamePlayerID,
				logging.FieldRemoteAddr, remoteAddr,
				"lane", candidate.Lane,
			)
		}) {
			return false
		}
		wire := realtime.WireLanePacket(candidate)
		logging.Network.Debug("lane protocol gameplay wire packet written",
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			logging.FieldRemoteAddr, remoteAddr,
			"wire_type", fmt.Sprint(wire["type"]),
			"candidate_lane", candidate.Lane,
			"candidate_kind", candidate.Kind,
			"wire_lane", fmt.Sprint(wire["lane"]),
			"sequence", fmt.Sprint(wire["sequence"]),
			"baseline_id", fmt.Sprint(wire["baseline_id"]),
			"snapshot_id", fmt.Sprint(wire["snapshot_id"]),
			"snapshot_kind", fmt.Sprint(wire["snapshot_kind"]),
			"encoded_bytes", len(encodedPacket),
		)
		if candidate.Kind == realtime.RealtimeLaneCandidateKindEventBatch {
			if drained := drainActiveEventBatchAfterWrite(session.room.GameInstance(), session.currentGamePlayerID, result.EventBatchEventIDs); len(drained) > 0 {
				drainedEventCount += len(drained)
			}
		}

		if metadata, ok := realtime.CandidateMetadata(candidate, session.realtimeState); ok {
			persistedMetadata := realtime.AdvanceMetadataForSuccessfulWrite(candidate.Lane, metadata)
			session.realtimeState.UpdateLane(candidate.Lane, persistedMetadata)
			if projection, ok := realtime.CandidateProjection(candidate); ok {
				session.realtimeState.StoreBaselineProjection(candidate.Lane, projection)
			}
			if metadata.IsFinalChunk && candidate.Kind == realtime.RealtimeLaneCandidateKindFull {
				session.realtimeState.MarkBaselineReady(candidate.Lane)
			}
		}
	}

	packetmetrics.LogSentLaneMetrics(result.MetricSummaries, session.currentRoomID, session.currentGamePlayerID, remoteAddr)
	logging.Network.Debug("lane protocol gameplay written",
		logging.FieldRoomID, session.currentRoomID,
		logging.FieldPlayerID, session.currentGamePlayerID,
		logging.FieldRemoteAddr, remoteAddr,
		"lane_packet_families", lanePacketFamilySummary(result.MetricSummaries),
		"baseline_full_count", countLaneCandidateKinds(result.SelectedCandidates, realtime.RealtimeLaneCandidateKindFull),
		"baseline_chunk_count", 0,
		"delta_blocked_count", len(result.SendPlan.Deferred),
		"event_batch_written", len(result.EventBatchEventIDs) > 0,
		"event_batch_drained_count", drainedEventCount,
		"candidate_count", len(result.Candidates),
		"included_count", len(result.SelectedCandidates),
		"packet_count", len(result.MetricSummaries),
		"encoded_bytes", result.TotalEncodedBytes,
	)
	return true
}

func countLaneCandidateKinds(candidates []realtime.RealtimeLaneCandidate, kind realtime.RealtimeLaneCandidateKind) int {
	count := 0
	for _, candidate := range candidates {
		if candidate.Kind == kind {
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
