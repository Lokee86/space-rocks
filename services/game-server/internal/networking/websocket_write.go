package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/services/game-server/internal/constants"
	game "github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/outbound"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

const debugStatusWriteIntervalTicks = 8

var canSendDebugShapeCatalog = outbound.CanSendDebugShapeCatalog
var buildDebugShapeCatalogResponse = outbound.BuildDebugShapeCatalogResponse

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

	if !maybeWriteDebugShapeCatalog(session, context, remoteAddr) {
		return false
	}

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
	session.realtimeState = result.SessionState
	transport := session.webRTCTransportSnapshot()
	for _, encoded := range result.EncodedLanePackets {
		candidate := encoded.Candidate
		encodedPacket := encoded.Encoded
		if len(encodedPacket) == 0 {
			continue
		}
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
					"lane":         string(candidate.Lane()),
					"transport":    "webrtc",
				},
			})
			return false
		}
		if transport == nil {
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
					"error_code":   "webrtc_transport_missing",
					"failure_mode": "webrtc_transport_missing",
					"lane":         string(candidate.Lane()),
					"transport":    "webrtc",
				},
			})
			continue
		}
		if !transport.Ready() {
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
					"error_code":   "webrtc_transport_not_ready",
					"failure_mode": "webrtc_transport_not_ready",
					"lane":         string(candidate.Lane()),
					"transport":    "webrtc",
				},
			})
			continue
		}
		channelLabel, ok := webRTCGameplayChannelLabelForLane(string(candidate.Lane()))
		if !ok {
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
					"error_code":   "webrtc_lane_channel_missing",
					"failure_mode": "webrtc_lane_channel_missing",
					"lane":         string(candidate.Lane()),
					"transport":    "webrtc",
				},
			})
			return false
		}
		if !session.sessionContextMatches(context) || !context.Room.GameplayContextMatches(gameplayContext) {
			return true
		}
		if err := transport.SendEncodedLaneJSON(string(candidate.Lane()), encodedPacket); err != nil {
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
					"lane":         string(candidate.Lane()),
					"transport":    "webrtc",
					"channel":      channelLabel,
				},
			})
			return false
		}
		if candidate.Kind() == realtime.RealtimeLaneCandidateKindEventBatch {
			drainActiveEventBatchAfterWrite(gameplayContext.Game, context.GamePlayerID, result.EventBatchEventIDs)
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

func maybeWriteDebugShapeCatalog(session *webSocketSession, context SessionContext, remoteAddr string) bool {
	if session == nil || context.Room == nil {
		return true
	}
	if session.debugShapeCatalogSentFor(context.RoomID) || !canSendDebugShapeCatalog(context.Room) {
		return true
	}
	response, ok := buildDebugShapeCatalogResponse(context.Room, context.RoomID)
	if !ok || !session.sessionContextMatches(context) {
		return true
	}
	if !outbound.WriteServerMessage(session.conn, response, func(err error) {
		logWebSocketWriteClose(err, session.connectionTraceID, session.sessionID, context.RoomID, context.GamePlayerID)
	}) {
		return false
	}
	session.markDebugShapeCatalogSent(context)
	return true
}
