package outbound

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func CanSendGameplayPresentationState(room *rooms.Room) bool {
	return room != nil &&
		room.Game != nil &&
		(room.State == rooms.RoomStateInGame || room.State == rooms.RoomStateGameOver)
}

func BuildGameplayPresentationStateResponse(room *rooms.Room, playerID string, roomID string, remoteAddr string) ([]byte, bool) {
	statePacket := room.Game.StatePacket(playerID)
	statePacket.ServerSentMsec = int(time.Now().UnixMilli())

	payload := any(statePacket)
	if devtools.Enabled() {
		payload = devtools.WrapStatePacket(
			statePacket,
			devtools.StatusFor(room.Game, playerID),
			devtools.StatusesForAllPlayers(room.Game),
		)
	}

	response, err := packetcodec.Encode(payload)
	if err != nil {
		logging.Network.Error("state packet encode failed", err,
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return nil, false
	}

	return response, true
}
