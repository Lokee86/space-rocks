package inbound

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

type telemetrySession interface {
	CurrentRoomID() string
	CurrentGamePlayerID() string
	SessionID() string
	OutboundMessages() chan<- []byte
}

func HandleTelemetryPacket(session telemetrySession, remoteAddr string, packet game.ClientPacket) bool {
	if packet.Type != game.PacketTypeTelemetryPing {
		return false
	}

	serverReceivedMsec := time.Now().UnixMilli()
	pong := game.ClientPacket{
		Type:               game.PacketTypeTelemetryPong,
		Sequence:           packet.Sequence,
		ClientSentMsec:     packet.ClientSentMsec,
		ServerReceivedMsec: int(serverReceivedMsec),
	}
	pong.ServerSentMsec = int(time.Now().UnixMilli())
	response, err := packetcodec.Encode(pong)
	if err != nil {
			logging.Network.Warn("websocket telemetry pong encode failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.CurrentRoomID(),
			logging.FieldPlayerID, session.CurrentGamePlayerID(),
			"session_id", session.SessionID(),
			logging.FieldRemoteAddr, remoteAddr,
		)
		return true
	}
	session.OutboundMessages() <- response
	return true
}
