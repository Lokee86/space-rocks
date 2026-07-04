package inbound

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

type webRTCSession interface {
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
			logging.Network.Warn("websocket webrtc offer decode failed",
				logging.FieldError, err,
				"packet_type", envelope.Type,
				logging.FieldRemoteAddr, remoteAddr,
			)
			return true
		}
		session.HandleWebRTCOffer(packet.DescriptionType, packet.Sdp)
		return true
	case game.PacketTypeWebrtcIceCandidate:
		var packet game.WebRTCIceCandidate
		if err := packetcodec.Decode(msg, &packet); err != nil {
			logging.Network.Warn("websocket webrtc ice candidate decode failed",
				logging.FieldError, err,
				"packet_type", envelope.Type,
				logging.FieldRemoteAddr, remoteAddr,
			)
			return true
		}
		session.HandleWebRTCIceCandidate(packet.Media, packet.Index, packet.Name)
		return true
	case game.PacketTypeWebrtcSmoke:
		var packet game.WebRTCSmoke
		if err := packetcodec.Decode(msg, &packet); err != nil {
			logging.Network.Warn("websocket webrtc smoke decode failed",
				logging.FieldError, err,
				"packet_type", envelope.Type,
				logging.FieldRemoteAddr, remoteAddr,
			)
			return true
		}
		session.HandleWebRTCSmoke(packet.SmokeID, packet.Origin, packet.Message)
		return true
	case game.PacketTypeWebrtcFailed:
		var packet game.WebRTCFailed
		if err := packetcodec.Decode(msg, &packet); err != nil {
			logging.Network.Warn("websocket webrtc failed decode failed",
				logging.FieldError, err,
				"packet_type", envelope.Type,
				logging.FieldRemoteAddr, remoteAddr,
			)
			return true
		}
		session.HandleWebRTCFailed(packet.ErrorCode, packet.Message)
		return true
	default:
		return false
	}
}

