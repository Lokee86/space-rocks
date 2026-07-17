package inbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

type webRTCSession interface {
	CurrentSessionContext() SessionContext
	SessionID() string
	ConnectionTraceID() string
	HandleWebRTCOffer(descriptionType string, sdp string)
	HandleWebRTCIceCandidate(media string, index int, name string)
	HandleWebRTCSmoke(smokeID string, origin string, message string)
	HandleWebRTCFailed(errorCode string, message string)
}

func HandleWebRTCSignalingPacket(session webRTCSession, remoteAddr string, msg []byte, envelope ClientPacketEnvelope) bool {
	switch envelope.Type {
	case game.PacketTypeWebrtcOffer:
		var packet game.WebRTCOffer
		if err := packetcodec.Decode(msg, &packet); err != nil {
			emitPacketDecodeFailure(session, envelope.Type, "webrtc_offer_decode_failed")
			return true
		}
		session.HandleWebRTCOffer(packet.DescriptionType, packet.Sdp)
		return true
	case game.PacketTypeWebrtcIceCandidate:
		var packet game.WebRTCIceCandidate
		if err := packetcodec.Decode(msg, &packet); err != nil {
			emitPacketDecodeFailure(session, envelope.Type, "webrtc_ice_candidate_decode_failed")
			return true
		}
		session.HandleWebRTCIceCandidate(packet.Media, packet.Index, packet.Name)
		return true
	case game.PacketTypeWebrtcSmoke:
		var packet game.WebRTCSmoke
		if err := packetcodec.Decode(msg, &packet); err != nil {
			emitPacketDecodeFailure(session, envelope.Type, "webrtc_smoke_decode_failed")
			return true
		}
		session.HandleWebRTCSmoke(packet.SmokeID, packet.Origin, packet.Message)
		return true
	case game.PacketTypeWebrtcFailed:
		var packet game.WebRTCFailed
		if err := packetcodec.Decode(msg, &packet); err != nil {
			emitPacketDecodeFailure(session, envelope.Type, "webrtc_failed_decode_failed")
			return true
		}
		session.HandleWebRTCFailed(packet.ErrorCode, packet.Message)
		return true
	default:
		return false
	}
}

func emitPacketDecodeFailure(session webRTCSession, packetType string, failureCode string) {
	context := session.CurrentSessionContext()
	logging.Emit(observability.Request{
		Event: observability.EventNamePacketDecodeFailed,
		Context: observability.Context{
			TraceID:    session.ConnectionTraceID(),
			SessionID:  session.SessionID(),
			RoomID:     context.RoomID,
			PlayerID:   context.GamePlayerID,
			PacketType: packetType,
		},
		Fields: observability.Fields{
			"error_code":   failureCode,
			"failure_mode": failureCode,
		},
	})
}
