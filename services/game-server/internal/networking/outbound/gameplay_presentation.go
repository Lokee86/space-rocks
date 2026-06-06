package outbound

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func CanSendGameplayPresentationState(room *rooms.Room) bool {
	return room != nil &&
		room.GameInstance() != nil &&
		(room.State == rooms.RoomStateInGame || room.State == rooms.RoomStateGameOver)
}

func BuildGameplayPresentationStateResponse(room *rooms.Room, playerID string, roomID string, remoteAddr string) ([]byte, bool) {
	gameInstance := room.GameInstance()
	statePacket := gameInstance.StatePacket(playerID)
	statePacket.ServerSentMsec = int(time.Now().UnixMilli())

	response, err := packetcodec.Encode(statePacket)
	if err != nil {
		logging.Network.Error("state packet encode failed", err,
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return nil, false
	}
	logGameplayPresentationPacketSize(len(response), roomID, playerID, remoteAddr)

	return response, true
}
