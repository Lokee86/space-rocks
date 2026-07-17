package outbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
	"github.com/google/uuid"
)

func CanSendDebugShapeCatalog(room *rooms.Room) bool {
	if room == nil {
		return false
	}
	context := room.GameplayContext()
	return room != nil &&
		context.Game != nil &&
		devtools.Enabled() &&
		(context.State == rooms.RoomStateInGame || context.State == rooms.RoomStateGameOver)
}

func BuildDebugShapeCatalogResponse(room *rooms.Room, roomID string) ([]byte, bool) {
	catalog, err := physics.LoadCollisionShapeCatalog()
	if err != nil {
		logging.Emit(observability.Request{
			Event:   observability.EventNameRuntimeAssetLoadFailed,
			Context: observability.Context{RoomID: roomID},
			Fields: observability.Fields{
				"error_code":   "collision_shape_catalog_load_failed",
				"failure_mode": "collision_shape_catalog_load_failed",
			},
		})
		return nil, false
	}

	responsePacket := devtools.DebugShapeCatalogPacket{
		Type:   "debug_shape_catalog",
		Shapes: devtools.BuildShapeCatalog(catalog),
	}

	response, err := packetcodec.Encode(responsePacket)
	if err != nil {
		logging.Emit(observability.Request{
			Event: observability.EventNameOutboundPacketEncodeFailed,
			Context: observability.Context{
				TraceID:    uuid.NewString(),
				RoomID:     roomID,
				PacketType: "debug_shape_catalog",
			},
			Fields: observability.Fields{
				"error_code":   "debug_shape_catalog_encode_failed",
				"failure_mode": "debug_shape_catalog_encode_failed",
			},
		})
		return nil, false
	}

	return response, true
}
