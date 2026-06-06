package outbound

import (
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func CanSendDebugStatus(room *rooms.Room) bool {
	return room != nil &&
		room.GameInstance() != nil &&
		devtools.Enabled() &&
		(room.State == rooms.RoomStateInGame || room.State == rooms.RoomStateGameOver)
}

func BuildDebugStatusResponse(room *rooms.Room, playerID string, roomID string, remoteAddr string) ([]byte, bool) {
	gameInstance := room.GameInstance()
	responsePacket := devtools.DebugStatusPacket{
		Type:          devtools.PacketTypeDebugStatus,
		DebugStatus:   devtools.StatusFor(gameInstance, playerID),
		DebugStatuses: devtools.StatusesForAllPlayers(gameInstance),
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
