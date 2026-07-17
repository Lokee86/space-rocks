package outbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
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

func BuildDebugStatusResponse(room *rooms.Room, playerID string, roomID string) ([]byte, bool) {
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
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:    uuid.NewString(),
				RoomID:     roomID,
				PlayerID:   playerID,
				PacketType: "debug_status",
			},
			Fields: observability.Fields{
				"error_code":   "debug_status_encode_failed",
				"failure_mode": "debug_status_encode_failed",
			},
		})
		return nil, false
	}

	return response, true
}
