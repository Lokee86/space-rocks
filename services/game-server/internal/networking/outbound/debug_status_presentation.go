package outbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

func CanSendDebugStatus(room *rooms.Room) bool {
	if room == nil {
		return false
	}
	context := room.GameplayContext()
	return context.Game != nil &&
		devtools.Enabled() &&
		(context.State == rooms.RoomStateInGame || context.State == rooms.RoomStateGameOver)
}

func BuildDebugStatusPacket(room *rooms.Room, playerID string, requestID string) (devtools.DebugStatusPacket, bool) {
	if !CanSendDebugStatus(room) {
		return devtools.DebugStatusPacket{}, false
	}
	context := room.GameplayContext()
	control := game.NewControl(context.Game)
	controller := devtools.NewController(devtools.Dependencies{Target: control})
	return devtools.DebugStatusPacket{
		Type:          devtools.PacketTypeDebugStatus,
		RequestID:     requestID,
		DebugStatus:   controller.StatusFor(playerID),
		DebugStatuses: controller.StatusesForAllPlayers(),
	}, true
}
