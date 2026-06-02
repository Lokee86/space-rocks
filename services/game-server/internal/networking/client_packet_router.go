package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

func handleClientPacket(session *webSocketSession, remoteAddr string, msg []byte, envelope clientPacketEnvelope) {
	var packet game.ClientPacket

	if handleSimpleDevtoolsPacket(session, remoteAddr, msg, envelope) {
		return
	}
	if handlePlacementDevtoolsPacket(session, remoteAddr, msg, envelope) {
		return
	}
	if handleRemainingDevtoolsPacket(session, remoteAddr, msg, envelope) {
		return
	}

	if err := packetcodec.Decode(msg, &packet); err != nil {
		logging.Network.Warn("websocket packet decode failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return
	}

	if handleTelemetryPacket(session, remoteAddr, packet) {
		return
	}
	if handleLobbyPacket(session, packet) {
		return
	}
	if handleGameplayPacket(session, packet) {
		return
	}
}
