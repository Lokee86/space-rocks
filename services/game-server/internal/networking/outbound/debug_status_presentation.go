package outbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func CanSendDebugStatus(room *rooms.Room) bool {
	if room == nil {
		return false
	}
	context := room.GameplayContext()
	return room != nil &&
		context.Game != nil &&
		devtools.Enabled() &&
		(context.State == rooms.RoomStateInGame || context.State == rooms.RoomStateGameOver)
}

func BuildDebugStatusResponse(room *rooms.Room, playerID string, roomID string, remoteAddr string) ([]byte, bool) {
	if room == nil {
		return nil, false
	}
	context := room.GameplayContext()
	if context.Game == nil {
		return nil, false
	}
	control := game.NewControl(context.Game)
	controller := devtools.NewController(devtools.Dependencies{Target: control})
	responsePacket := devtools.DebugStatusPacket{
		Type:          devtools.PacketTypeDebugStatus,
		DebugStatus:   controller.StatusFor(playerID),
		DebugStatuses: controller.StatusesForAllPlayers(),
	}

	response, err := packetcodec.Encode(responsePacket)
	if err != nil {
		logging.Network.Error("debug status packet encode failed", err,
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return nil, false
	}

	return response, true
}
