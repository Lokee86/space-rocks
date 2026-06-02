package networking

import (
	"time"

	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func canSendGameplayPresentationState(room *rooms.Room) bool {
	return room != nil &&
		room.Game != nil &&
		(room.State == rooms.RoomStateInGame || room.State == rooms.RoomStateGameOver)
}

func buildGameplayPresentationStateResponse(session *webSocketSession, remoteAddr string) ([]byte, bool) {
	statePacket := session.room.Game.StatePacket(session.currentGamePlayerID)
	statePacket.ServerSentMsec = int(time.Now().UnixMilli())

	payload := any(statePacket)
	if devtools.Enabled() {
		payload = devtools.WrapStatePacket(
			statePacket,
			devtools.StatusFor(session.room.Game, session.currentGamePlayerID),
			devtools.StatusesForAllPlayers(session.room.Game),
		)
	}

	response, err := packetcodec.Encode(payload)
	if err != nil {
		logging.Network.Error("state packet encode failed", err,
			logging.FieldRoomID, session.currentRoomID,
			logging.FieldPlayerID, session.currentGamePlayerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return nil, false
	}

	return response, true
}
