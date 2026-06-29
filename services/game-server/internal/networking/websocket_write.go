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

var loggedEventBatchWriteIDs = map[string]bool{}

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
	for _, candidate := range result.Candidates {
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
		if candidate.Kind == realtime.RealtimeLaneCandidateKindEventBatch {
			encodedEventCount := len(result.EventBatchEventIDs)
			eventTypes := make([]string, 0, encodedEventCount)
			eventIDs := make([]string, 0, encodedEventCount)
			shipDeathFound := false
			if eventBatch, ok := candidate.Full.(map[string]any); ok {
				if events, ok := eventBatch["events"].([]any); ok {
					for idx, event := range events {
						eventID := fmt.Sprint(result.EventBatchEventIDs[idx])
						eventIDs = append(eventIDs, eventID)
						if eventMap, ok := event.(map[string]any); ok {
							eventType := fmt.Sprint(eventMap["type"])
							eventTypes = append(eventTypes, eventType)
							if eventType == "ship_death" {
								shipDeathFound = true
								logging.Network.Debug("ship death present in active event batch",
									logging.FieldRoomID, session.currentRoomID,
									logging.FieldPlayerID, session.currentGamePlayerID,
									"event_id", eventID,
									"event_type", eventType,
									"event_player_id", fmt.Sprint(eventMap["player_id"]),
									"lives", fmt.Sprint(eventMap["lives"]),
									"respawn_delay", fmt.Sprint(eventMap["respawn_delay"]),
								)
							}
						}
					}
				}
			}
			if drained := drainActiveEventBatchAfterWrite(session.room.GameInstance(), session.currentGamePlayerID, result.EventBatchEventIDs); len(drained) > 0 {
				drainedEventCount += len(drained)
				if shipDeathFound {
					logging.Network.Debug("lane protocol event batch drained",
						logging.FieldRoomID, session.currentRoomID,
						logging.FieldPlayerID, session.currentGamePlayerID,
						logging.FieldRemoteAddr, remoteAddr,
						"event_count", len(drained),
					)
				}
			}
			if shipDeathFound && encodedEventCount > 0 {
				batchID := ""
				if eventBatch, ok := candidate.Full.(map[string]any); ok {
					batchID = fmt.Sprint(eventBatch["batch_id"])
				}
				if !loggedEventBatchWriteIDs[batchID] {
					loggedEventBatchWriteIDs[batchID] = true
					logging.Network.Debug("lane protocol event batch written",
						logging.FieldRoomID, session.currentRoomID,
						logging.FieldPlayerID, session.currentGamePlayerID,
						logging.FieldRemoteAddr, remoteAddr,
						"batch_id", batchID,
						"event_count", encodedEventCount,
						"event_ids", strings.Join(eventIDs, ","),
						"event_types", strings.Join(eventTypes, ","),
						"event_batch_drained_count", drainedEventCount,
					)
				}
			}
		}
		if metadata, ok := realtime.CandidateMetadata(candidate, session.realtimeState); ok {
			persistedMetadata := realtime.AdvanceMetadataForSuccessfulWrite(candidate.Lane, metadata)
			session.realtimeState.UpdateLane(candidate.Lane, persistedMetadata)
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
		"baseline_full_count", countLaneCandidateKinds(result.Candidates, realtime.RealtimeLaneCandidateKindFull),
		"baseline_chunk_count", 0,
		"delta_blocked_count", len(result.SendPlan.Deferred),
		"event_batch_written", len(result.EventBatchEventIDs) > 0,
		"event_batch_drained_count", drainedEventCount,
		"candidate_count", len(result.Candidates),
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
