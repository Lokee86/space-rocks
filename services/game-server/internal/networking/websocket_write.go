package networking

import (
	"errors"
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/outbound"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func writeServerMessages(session *webSocketSession, remoteAddr string, readErr <-chan error) {
	ticker := time.NewTicker(time.Second / time.Duration(constants.ServerTickRate))
	defer ticker.Stop()

	for {
		select {
		case err := <-readErr:
			context := session.sessionContext()
			logWebSocketReadClose(err, session.connectionTraceID, session.sessionID, context.RoomID, context.GamePlayerID)
			return
		case request := <-session.resyncRequests:
			if !writeResyncRequiredAndApply(session, request, remoteAddr) {
				return
			}
		case message := <-session.outbound:
			context := session.sessionContext()
			if !outbound.WriteServerMessage(session.conn, message, func(err error) {
				logWebSocketWriteClose(err, session.connectionTraceID, session.sessionID, context.RoomID, context.GamePlayerID)
			}) {
				return
			}
		case <-ticker.C:
			session.writeToolingProtocolMessage()
			if !writeGameplayLaneProtocolMessage(session, remoteAddr) {
				return
			}
		}
	}
}

func writeGameplayLaneProtocolMessage(session *webSocketSession, remoteAddr string) bool {
	context := session.sessionContext()
	if context.Room == nil || context.GamePlayerID == "" {
		return true
	}
	gameplayContext := context.Room.GameplayContext()
	if gameplayContext.Game == nil {
		return true
	}
	resetRealtimeStateForContext(session, context, gameplayContext.MatchID)

	result, err := realtime.BuildActiveRealtimeResultForGame(gameplayContext.Game, context.GamePlayerID, session.realtimeState)
	if err != nil {
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:   session.connectionTraceID,
				SessionID: session.sessionID,
				RoomID:    context.RoomID,
				PlayerID:  context.GamePlayerID,
				MatchID:   gameplayContext.MatchID,
			},
			Fields: observability.Fields{
				"error_code":   "realtime_payload_build_failed",
				"failure_mode": "realtime_payload_build_failed",
			},
		})
		return false
	}

	if !session.sessionContextMatches(context) || !context.Room.GameplayContextMatches(gameplayContext) {
		return true
	}
	nextRealtimeState := result.SessionState
	transport := session.webRTCTransportSnapshot()
	if transport == nil || !transport.Ready() {
		return true
	}

	reservedBytesByLane := make(map[string]uint64)
	for _, encoded := range result.EncodedLanePackets {
		candidate := encoded.Candidate
		if len(encoded.Encoded) == 0 {
			continue
		}
		lane := string(candidate.Lane())
		if candidate.Lane() == realtime.LaneControl {
			logging.Emit(observability.Request{
				Event: observability.EventNamePacketRouteFailed,
				Context: observability.Context{
					TraceID:    session.connectionTraceID,
					SessionID:  session.sessionID,
					RoomID:     context.RoomID,
					PlayerID:   context.GamePlayerID,
					MatchID:    gameplayContext.MatchID,
					PacketType: candidate.PacketFamily(),
				},
				Fields: observability.Fields{
					"error_code":   "control_lane_wrong_transport",
					"failure_mode": "control_lane_wrong_transport",
					"lane":         lane,
					"transport":    "webrtc",
				},
			})
			return false
		}
		channelLabel, ok := webRTCChannelLabelForLane(lane)
		if !ok {
			return false
		}
		bufferedBytes, ok := transport.BufferedAmountForLane(lane)
		if !ok {
			return false
		}
		reservedBytes := reservedBytesByLane[lane]
		payloadBytes := uint64(len(encoded.Encoded))
		if bufferedBytes >= webRTCMaxBufferedAmountBytes || reservedBytes+payloadBytes > webRTCMaxBufferedAmountBytes-bufferedBytes {
			if session.shouldLogWebRTCBackpressure(lane) {
				logging.Emit(observability.Request{
					Event: observability.EventNamePacketRouteFailed,
					Context: observability.Context{
						TraceID:    session.connectionTraceID,
						SessionID:  session.sessionID,
						RoomID:     context.RoomID,
						PlayerID:   context.GamePlayerID,
						MatchID:    gameplayContext.MatchID,
						PacketType: candidate.PacketFamily(),
					},
					Fields: observability.Fields{
						"error_code":     "webrtc_channel_backpressure",
						"failure_mode":   "webrtc_channel_backpressure",
						"lane":           lane,
						"transport":      "webrtc",
						"channel":        channelLabel,
						"buffered_bytes": bufferedBytes,
						"reserved_bytes": reservedBytes,
					},
				})
			}
			return true
		}
		reservedBytesByLane[lane] = reservedBytes + payloadBytes
	}

	sentEventBatch := false
	for _, encoded := range result.EncodedLanePackets {
		candidate := encoded.Candidate
		encodedPacket := encoded.Encoded
		if len(encodedPacket) == 0 {
			continue
		}
		if !session.sessionContextMatches(context) || !context.Room.GameplayContextMatches(gameplayContext) {
			return true
		}
		lane := string(candidate.Lane())
		channelLabel, _ := webRTCChannelLabelForLane(lane)
		if err := transport.SendEncodedLaneJSON(lane, encodedPacket); err != nil {
			if errors.Is(err, ErrWebRTCChannelBackpressure) {
				if session.shouldLogWebRTCBackpressure(lane) {
					bufferedBytes, _ := transport.BufferedAmountForLane(lane)
					logging.Emit(observability.Request{
						Event: observability.EventNamePacketRouteFailed,
						Context: observability.Context{
							TraceID:    session.connectionTraceID,
							SessionID:  session.sessionID,
							RoomID:     context.RoomID,
							PlayerID:   context.GamePlayerID,
							MatchID:    gameplayContext.MatchID,
							PacketType: candidate.PacketFamily(),
						},
						Fields: observability.Fields{
							"error_code":     "webrtc_channel_backpressure",
							"failure_mode":   "webrtc_channel_backpressure",
							"lane":           lane,
							"transport":      "webrtc",
							"channel":        channelLabel,
							"buffered_bytes": bufferedBytes,
						},
					})
				}
				return true
			}
			logging.Emit(observability.Request{
				Event: observability.EventNameGameServerWriteFailed,
				Context: observability.Context{
					TraceID:    session.connectionTraceID,
					SessionID:  session.sessionID,
					RoomID:     context.RoomID,
					PlayerID:   context.GamePlayerID,
					MatchID:    gameplayContext.MatchID,
					PacketType: candidate.PacketFamily(),
				},
				Fields: observability.Fields{
					"error_code":   "webrtc_lane_write_failed",
					"failure_mode": "webrtc_lane_write_failed",
					"lane":         lane,
					"transport":    "webrtc",
					"channel":      channelLabel,
				},
			})
			return false
		}
		session.observePacketWrite(lane, candidate.PacketFamily(), len(encodedPacket))
		if candidate.Kind() == realtime.RealtimeLaneCandidateKindEventBatch {
			sentEventBatch = true
		}
		if metadata, ok := candidate.Metadata(); ok {
			persistedMetadata := realtime.AdvanceMetadataForSuccessfulWrite(candidate.Lane(), metadata)
			nextRealtimeState.UpdateLane(candidate.Lane(), persistedMetadata)
			if projection, ok := realtime.CandidateProjection(candidate); ok {
				nextRealtimeState.StoreBaselineProjection(candidate.Lane(), projection)
			}
			if metadata.IsFinalChunk && candidate.Kind() == realtime.RealtimeLaneCandidateKindFull {
				nextRealtimeState.MarkBaselineReady(candidate.Lane())
			}
		}
	}

	session.realtimeState = nextRealtimeState
	if sentEventBatch {
		drainActiveEventBatchAfterWrite(gameplayContext.Game, context.GamePlayerID, result.EventBatchEventIDs)
	}
	return true
}

func resetRealtimeStateForContext(session *webSocketSession, context SessionContext, matchID string) {
	if session == nil || context.Room == nil || context.GamePlayerID == "" {
		return
	}
	if !session.realtimeState.IdentityMatches(context.GamePlayerID, matchID) {
		session.realtimeState = realtime.NewRealtimeSessionState(context.GamePlayerID, matchID)
	}
}

func drainActiveEventBatchAfterWrite(gameInstance *game.Game, playerID string, eventIDs []string) []game.PendingPresentationEvent {
	if gameInstance == nil || len(eventIDs) == 0 {
		return nil
	}
	return gameInstance.DrainPendingPresentationEvents(playerID, eventIDs...)
}
