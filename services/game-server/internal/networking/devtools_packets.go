package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/networking/inbound"
)

func handleSimpleDevtoolsPacket(session *webSocketSession, remoteAddr string, msg []byte, envelope inbound.ClientPacketEnvelope) bool {
	if !isSimpleDevtoolsPacketType(envelope.Type) {
		return false
	}
	return handleDevtoolsCommandPacket(session, remoteAddr, msg)
}

func isSimpleDevtoolsPacketType(packetType string) bool {
	switch packetType {
	case devtools.PacketTypeToggleDebugInvincible,
		devtools.PacketTypeToggleDebugInfiniteLives,
		devtools.PacketTypeToggleDebugFreezeWorld,
		devtools.PacketTypeToggleDebugFreezePlayer,
		devtools.PacketTypeDebugKillPlayer,
		devtools.PacketTypeDebugSetScore,
		devtools.PacketTypeDebugAddScore,
		devtools.PacketTypeDebugSetLives,
		devtools.PacketTypeDebugAddLives,
		devtools.PacketTypeDebugClearBullets,
		devtools.PacketTypeDebugClearAsteroids:
		return true
	default:
		return false
	}
}

func handlePlacementDevtoolsPacket(session *webSocketSession, remoteAddr string, msg []byte, envelope inbound.ClientPacketEnvelope) bool {
	if envelope.Type != devtools.PacketTypeDebugSpawnEntity {
		return false
	}
	return handleDevtoolsCommandPacket(session, remoteAddr, msg)
}

func handleRemainingDevtoolsPacket(session *webSocketSession, remoteAddr string, msg []byte, envelope inbound.ClientPacketEnvelope) bool {
	if !isRemainingDevtoolsPacketType(envelope.Type) {
		return false
	}
	return handleDevtoolsCommandPacket(session, remoteAddr, msg)
}

func isRemainingDevtoolsPacketType(packetType string) bool {
	switch packetType {
	case devtools.PacketTypeDebugBeginContinuousBulletStream,
		devtools.PacketTypeDebugRespawnPlayer:
		return true
	default:
		return false
	}
}

func handleDevtoolsCommandPacket(session *webSocketSession, remoteAddr string, msg []byte) bool {
	if session.room == nil || session.currentGamePlayerID == "" {
		return true
	}

	var command devtools.DebugCommand
	if err := packetcodec.Decode(msg, &command); err != nil {
		logging.Network.Warn("websocket devtools command decode failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			"session_id", session.sessionID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return true
	}
	devtools.HandleCommand(session.room.Game, session.currentGamePlayerID, command)
	return true
}
