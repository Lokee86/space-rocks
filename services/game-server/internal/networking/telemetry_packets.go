package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

func handleTelemetryPacket(session *webSocketSession, remoteAddr string, packet game.ClientPacket) bool {
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
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return true
	}
	session.outbound <- response
	return true
}
