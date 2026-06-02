package inbound

import (
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

type devtoolsSession interface {
	CurrentRoom() *rooms.Room
	CurrentRoomID() string
	CurrentGamePlayerID() string
	SessionID() string
}

func HandleSimpleDevtoolsPacket(session devtoolsSession, remoteAddr string, msg []byte, envelope ClientPacketEnvelope) bool {
	if !isSimpleDevtoolsPacketType(envelope.Type) {
		return false
	}
	return handleDevtoolsCommandPacket(session, remoteAddr, msg)
}

func HandlePlacementDevtoolsPacket(session devtoolsSession, remoteAddr string, msg []byte, envelope ClientPacketEnvelope) bool {
	if envelope.Type != devtools.PacketTypeDebugSpawnEntity {
		return false
	}
	return handleDevtoolsCommandPacket(session, remoteAddr, msg)
}

func HandleRemainingDevtoolsPacket(session devtoolsSession, remoteAddr string, msg []byte, envelope ClientPacketEnvelope) bool {
	if !isRemainingDevtoolsPacketType(envelope.Type) {
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

func isRemainingDevtoolsPacketType(packetType string) bool {
	switch packetType {
	case devtools.PacketTypeDebugBeginContinuousBulletStream,
		devtools.PacketTypeDebugRespawnPlayer:
		return true
	default:
		return false
	}
}

func handleDevtoolsCommandPacket(session devtoolsSession, remoteAddr string, msg []byte) bool {
	if session.CurrentRoom() == nil || session.CurrentGamePlayerID() == "" {
		return true
	}

	var command devtools.DebugCommand
	if err := packetcodec.Decode(msg, &command); err != nil {
		logging.Network.Warn("websocket devtools command decode failed",
			logging.FieldError, err,
			logging.FieldRoomID, session.CurrentRoomID(),
			logging.FieldPlayerID, session.CurrentGamePlayerID(),
			"session_id", session.SessionID(),
			logging.FieldRemoteAddr, remoteAddr,
		)
		return true
	}
	devtools.HandleCommand(session.CurrentRoom().Game, session.CurrentGamePlayerID(), command)
	return true
}
