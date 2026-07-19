package outbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func CanSendDebugShapeCatalog(room *rooms.Room) bool {
	if room == nil {
		return false
	}
	context := room.GameplayContext()
	return context.Game != nil &&
		devtools.Enabled() &&
		(context.State == rooms.RoomStateInGame || context.State == rooms.RoomStateGameOver)
}

func BuildDebugShapeCatalogPacket(room *rooms.Room, roomID string, requestID string) (devtools.DebugShapeCatalogPacket, bool) {
	if !CanSendDebugShapeCatalog(room) {
		return devtools.DebugShapeCatalogPacket{}, false
	}
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
		return devtools.DebugShapeCatalogPacket{}, false
	}

	return devtools.DebugShapeCatalogPacket{
		Type:      devtools.PacketTypeDebugShapeCatalog,
		RequestID: requestID,
		Shapes:    devtools.BuildShapeCatalog(catalog),
	}, true
}
