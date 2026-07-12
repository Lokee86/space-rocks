package outbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/devtools"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
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

func BuildDebugShapeCatalogResponse(room *rooms.Room, roomID string, remoteAddr string) ([]byte, bool) {
	catalog, err := physics.LoadCollisionShapeCatalog()
	if err != nil {
		logging.Network.Error("debug shape catalog load failed", err,
			logging.FieldRoomID, roomID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return nil, false
	}

	responsePacket := devtools.DebugShapeCatalogPacket{
		Type:   "debug_shape_catalog",
		Shapes: devtools.BuildShapeCatalog(catalog),
	}

	response, err := packetcodec.Encode(responsePacket)
	if err != nil {
		logging.Network.Error("debug shape catalog packet encode failed", err,
			logging.FieldRoomID, roomID,
			logging.FieldRemoteAddr, remoteAddr,
		)
		return nil, false
	}

	return response, true
}
