package outbound

import (
	"github.com/Lokee86/space-rocks/server/internal/devtools"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

func CanSendDebugShapeCatalog(room *rooms.Room) bool {
	return room != nil &&
		room.GameInstance() != nil &&
		devtools.Enabled() &&
		(room.State == rooms.RoomStateInGame || room.State == rooms.RoomStateGameOver)
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
