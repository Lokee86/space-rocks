package outbound

import (
	"time"

	game "github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking/packetmetrics"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func CanSendGameplayPresentationState(room *rooms.Room) bool {
	return room != nil &&
		room.GameInstance() != nil &&
		(room.State == rooms.RoomStateInGame || room.State == rooms.RoomStateGameOver)
}

func BuildGameplayPresentationStateResponse(room *rooms.Room, playerID string, roomID string, remoteAddr string) ([]byte, packetmetrics.GameplayPresentationPacketMetrics, bool) {
	gameInstance := room.GameInstance()
	buildStarted := time.Now()
	statePacket := gameInstance.StatePacket(playerID)
	statePacket.ServerSentMsec = int(time.Now().UnixMilli())
	contributors := buildGameplayPacketContributors(string(room.State), gameInstance.GameplayPresentationSnapshot(playerID))
	buildDuration := time.Since(buildStarted)

	encodeStarted := time.Now()
	response, err := packetcodec.Encode(statePacket)
	encodeDuration := time.Since(encodeStarted)
	if err != nil {
		logging.Network.Error("state packet encode failed", err,
			logging.FieldRoomID, roomID,
			logging.FieldPlayerID, playerID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return nil, packetmetrics.GameplayPresentationPacketMetrics{}, false
	}

	metrics := packetmetrics.NewGameplayPresentationPacketMetrics(len(response), contributors, buildDuration, encodeDuration)
	packetmetrics.LogGameplayPresentationPacketSize(metrics, roomID, playerID, remoteAddr)

	return response, metrics, true
}

func buildGameplayPacketContributors(roomState string, snapshot game.GameplayPresentationSnapshot) packetmetrics.GameplayPacketContributors {
	return packetmetrics.GameplayPacketContributors{
		RoomState:      roomState,
		Players:        len(snapshot.Players),
		PlayerSessions: len(snapshot.PlayerSessions),
		PlayerLifecycle: len(snapshot.PlayerLifecycle),
		Asteroids:      len(snapshot.Asteroids),
		Bullets:        len(snapshot.Bullets),
		Pickups:        len(snapshot.Pickups),
		Enemies:        0,
		Events:         len(snapshot.PendingEvents),
		TotalAsteroids: snapshot.TotalAsteroids,
	}
}
