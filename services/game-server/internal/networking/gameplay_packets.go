package networking

import (
	targeting "github.com/Lokee86/space-rocks/server/internal/game/targeting"
	"github.com/Lokee86/space-rocks/server/internal/game"
)

func handleGameplayPacket(session *webSocketSession, packet game.ClientPacket) bool {
	if packet.Type != game.PacketTypeInput && packet.Type != game.PacketTypeRespawn {
		if session.room == nil || session.currentGamePlayerID == "" {
			return false
		}
		switch packet.Type {
		case game.PacketTypeSetTargetPlayerRequest:
			session.room.Game.SetPlayerTarget(session.currentGamePlayerID, packet.TargetPlayerID)
			return true
		case game.PacketTypeSelectTargetAtPositionRequest:
			session.room.Game.SelectTargetAtPosition(
				session.currentGamePlayerID,
				packet.X,
				packet.Y,
				targeting.TargetRef{
					Kind: targeting.TargetKind(packet.TargetKind),
					ID:   packet.TargetID,
				},
			)
			return true
		case game.PacketTypeClearTargetRequest:
			session.room.Game.ClearTarget(session.currentGamePlayerID)
			return true
		case game.PacketTypePauseRequest:
			session.room.Game.HandlePacket(session.currentGamePlayerID, packet)
			session.EnqueuePlayerPauseState()
			return true
		}
		return false
	}
	if session.room == nil || session.currentGamePlayerID == "" {
		return true
	}

	session.room.Game.HandlePacket(session.currentGamePlayerID, packet)
	return true
}
