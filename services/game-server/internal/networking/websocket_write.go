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

	result, err := realtime.BuildActiveRealtimeResultForGameView(
		gameplayContext.Game,
		context.GamePlayerID,
		session.viewTargetPlayerIDSnapshot(),
		session.realtimeState,
	)
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

	sentEventBatch := false
	for _, group := range groupEncodedRealtimeLanePackets(result.EncodedLanePackets) {
		blocked, valid := preflightRealtimeSendGroup(session, transport, group, context, gameplayContext.MatchID)
		if !valid {
			return false
		}
		if blocked {
			continue
		}

		groupSent := true
		for _, encoded := range group.Packets {
			candidate := encoded.Candidate
			if len(encoded.Encoded) == 0 {
				continue
			}
			if !session.sessionContextMatches(context) || !context.Room.GameplayContextMatches(gameplayContext) {
				session.realtimeState = nextRealtimeState
				return true
			}
			lane := string(candidate.Lane())
			channelLabel, _ := webRTCChannelLabelForLane(lane)
			if err := transport.SendEncodedLaneJSON(lane, encoded.Encoded); err != nil {
				if errors.Is(err, ErrWebRTCChannelBackpressure) {
					logRealtimeBackpressure(session, transport, candidate, lane, channelLabel, context, gameplayContext.MatchID, 0)
					groupSent = false
					break
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
				session.realtimeState = nextRealtimeState
				return false
			}
			session.observePacketWrite(lane, candidate.PacketFamily(), len(encoded.Encoded))
		}
		if !groupSent {
			continue
		}

		for _, encoded := range group.Packets {
			realtime.CommitSuccessfulCandidate(&nextRealtimeState, encoded.Candidate)
			if encoded.Candidate.Kind() == realtime.RealtimeLaneCandidateKindEventBatch {
				sentEventBatch = true
			}
		}
		session.realtimeState = nextRealtimeState
	}

	// Cadence and cohort state advance even when an individual hot lane is
	// backpressured. Only that lane's metadata and projection remain uncommitted.
	session.realtimeState = nextRealtimeState
	if sentEventBatch {
		drainActiveEventBatchAfterWrite(gameplayContext.Game, context.GamePlayerID, result.EventBatchEventIDs)
	}
	return true
}

type encodedRealtimeLaneGroup struct {
	Key     string
	Packets []realtime.EncodedRealtimeLanePacket
}

func groupEncodedRealtimeLanePackets(packets []realtime.EncodedRealtimeLanePacket) []encodedRealtimeLaneGroup {
	groups := make([]encodedRealtimeLaneGroup, 0, len(packets))
	indexes := make(map[string]int, len(packets))
	for _, packet := range packets {
		key := realtimeSendGroupKey(packet.Candidate)
		if index, ok := indexes[key]; ok {
			groups[index].Packets = append(groups[index].Packets, packet)
			continue
		}
		indexes[key] = len(groups)
		groups = append(groups, encodedRealtimeLaneGroup{Key: key, Packets: []realtime.EncodedRealtimeLanePacket{packet}})
	}
	return groups
}

func realtimeSendGroupKey(candidate realtime.RealtimeLaneCandidate) string {
	switch candidate.Lane() {
	case realtime.LaneWorld, realtime.LaneShipsLifecycle, realtime.LaneAsteroidsLifecycle, realtime.LaneBulletsLifecycle:
		return "world.reliable"
	default:
		return string(candidate.Lane())
	}
}

func preflightRealtimeSendGroup(session *webSocketSession, transport *WebRTCTransport, group encodedRealtimeLaneGroup, context SessionContext, matchID string) (bool, bool) {
	reservedBytesByLane := make(map[string]uint64)
	for _, encoded := range group.Packets {
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
					MatchID:    matchID,
					PacketType: candidate.PacketFamily(),
				},
				Fields: observability.Fields{
					"error_code":   "control_lane_wrong_transport",
					"failure_mode": "control_lane_wrong_transport",
					"lane":         lane,
					"transport":    "webrtc",
				},
			})
			return false, false
		}
		channelLabel, ok := webRTCChannelLabelForLane(lane)
		if !ok {
			return false, false
		}
		bufferedBytes, ok := transport.BufferedAmountForLane(lane)
		if !ok {
			return false, false
		}
		reservedBytes := reservedBytesByLane[lane]
		payloadBytes := uint64(len(encoded.Encoded))
		if bufferedBytes >= webRTCMaxBufferedAmountBytes || reservedBytes+payloadBytes > webRTCMaxBufferedAmountBytes-bufferedBytes {
			logRealtimeBackpressure(session, transport, candidate, lane, channelLabel, context, matchID, reservedBytes)
			return true, true
		}
		reservedBytesByLane[lane] = reservedBytes + payloadBytes
	}
	return false, true
}

func logRealtimeBackpressure(session *webSocketSession, transport *WebRTCTransport, candidate realtime.RealtimeLaneCandidate, lane string, channelLabel string, context SessionContext, matchID string, reservedBytes uint64) {
	if !session.shouldLogWebRTCBackpressure(lane) {
		return
	}
	bufferedBytes, _ := transport.BufferedAmountForLane(lane)
	logging.Emit(observability.Request{
		Event: observability.EventNamePacketRouteFailed,
		Context: observability.Context{
			TraceID:    session.connectionTraceID,
			SessionID:  session.sessionID,
			RoomID:     context.RoomID,
			PlayerID:   context.GamePlayerID,
			MatchID:    matchID,
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

func resetRealtimeStateForContext(session *webSocketSession, context SessionContext, matchID string) {
	if session == nil || context.Room == nil || context.GamePlayerID == "" {
		return
	}
	if !session.realtimeState.IdentityMatches(context.GamePlayerID, matchID) {
		session.realtimeState = realtime.NewRealtimeSessionState(context.GamePlayerID, matchID)
		session.clearViewTargetPlayerID()
	}
}

func drainActiveEventBatchAfterWrite(gameInstance *game.Game, playerID string, eventIDs []string) []game.PendingPresentationEvent {
	if gameInstance == nil || len(eventIDs) == 0 {
		return nil
	}
	return gameInstance.DrainPendingPresentationEvents(playerID, eventIDs...)
}
