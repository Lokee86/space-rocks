package inbound

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	targeting "github.com/Lokee86/space-rocks/server/internal/game/targeting"
	"github.com/Lokee86/space-rocks/server/internal/rooms"
)

type gameplaySession interface {
	CurrentRoom() *rooms.Room
	CurrentGamePlayerID() string
	EnqueuePlayerPauseState()
}

func HandleGameplayPacket(session gameplaySession, packet game.ClientPacket) bool {
	if packet.Type != game.PacketTypeInput && packet.Type != game.PacketTypeRespawn && packet.Type != game.PacketTypeClientConfig {
		room := session.CurrentRoom()
		if room == nil || session.CurrentGamePlayerID() == "" {
			return false
		}
		gameInstance := room.GameInstance()
		switch packet.Type {
		case game.PacketTypeSetTargetPlayerRequest:
			gameInstance.SetPlayerTarget(session.CurrentGamePlayerID(), packet.TargetPlayerID)
			return true
		case game.PacketTypeSelectTargetAtPositionRequest:
			gameInstance.SelectTargetAtPosition(
				session.CurrentGamePlayerID(),
				packet.X,
				packet.Y,
				targeting.TargetRef{
					Kind: targeting.TargetKind(packet.TargetKind),
					ID:   packet.TargetID,
				},
			)
			return true
		case game.PacketTypeClearTargetRequest:
			gameInstance.ClearTarget(session.CurrentGamePlayerID())
			return true
		case game.PacketTypePauseRequest:
			gameInstance.HandlePacket(session.CurrentGamePlayerID(), packet)
			session.EnqueuePlayerPauseState()
			return true
		}
		return false
	}
	room := session.CurrentRoom()
	if room == nil || session.CurrentGamePlayerID() == "" {
		return true
	}

	room.GameInstance().HandlePacket(session.CurrentGamePlayerID(), packet)
	return true
}
