package inbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
)

type devtoolsSession interface {
	CurrentSessionContext() SessionContext
	SessionID() string
}

func HandleSimpleDevtoolsPacket(session devtoolsSession, remoteAddr string, msg []byte, envelope ClientPacketEnvelope) bool {
	if !isSimpleDevtoolsPacketType(envelope.Type) {
		return false
	}
	return handleDevtoolsCommandPacket(session, remoteAddr, msg)
}

func HandlePlacementDevtoolsPacket(session devtoolsSession, remoteAddr string, msg []byte, envelope ClientPacketEnvelope) bool {
	switch envelope.Type {
	case devtools.PacketTypeDebugSpawnEntity, devtools.PacketTypeDebugSpawnPickup:
	default:
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
	case devtools.PacketTypeToggleDebugInvincible, devtools.PacketTypeToggleDebugInfiniteLives, devtools.PacketTypeToggleDebugFreezeWorld, devtools.PacketTypeToggleDebugFreezePlayer, devtools.PacketTypeDebugKillPlayer, devtools.PacketTypeDebugSetScore, devtools.PacketTypeDebugAddScore, devtools.PacketTypeDebugSetLives, devtools.PacketTypeDebugAddLives, devtools.PacketTypeDebugClearBullets, devtools.PacketTypeDebugClearAsteroids:
		return true
	default:
		return false
	}
}

func isRemainingDevtoolsPacketType(packetType string) bool {
	switch packetType {
	case devtools.PacketTypeDebugBeginContinuousBulletStream, devtools.PacketTypeDebugRespawnPlayer:
		return true
	default:
		return false
	}
}

func handleDevtoolsCommandPacket(session devtoolsSession, remoteAddr string, msg []byte) bool {
	context := session.CurrentSessionContext()
	if context.Room == nil || context.GamePlayerID == "" {
		return true
	}
	gameplayContext := context.Room.GameplayContext()
	if gameplayContext.Game == nil {
		return true
	}
	var command devtools.DebugCommand
	if err := packetcodec.Decode(msg, &command); err != nil {
		logging.Network.Warn("websocket devtools command decode failed", logging.FieldError, err, logging.FieldRoomID, context.RoomID, logging.FieldPlayerID, context.GamePlayerID, "session_id", session.SessionID(), logging.FieldRemoteAddr, remoteAddr)
		return true
	}
	control := game.NewControl(gameplayContext.Game)
	controller := devtools.NewController(devtools.Dependencies{Target: control})
	controller.HandleCommand(context.GamePlayerID, command)
	return true
}
