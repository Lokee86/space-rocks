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
		if session.CurrentRoom() == nil || session.CurrentGamePlayerID() == "" {
			return false
		}
		switch packet.Type {
		case game.PacketTypeSetTargetPlayerRequest:
			session.CurrentRoom().Game.SetPlayerTarget(session.CurrentGamePlayerID(), packet.TargetPlayerID)
			return true
		case game.PacketTypeSelectTargetAtPositionRequest:
			session.CurrentRoom().Game.SelectTargetAtPosition(
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
			session.CurrentRoom().Game.ClearTarget(session.CurrentGamePlayerID())
			return true
		case game.PacketTypePauseRequest:
			session.CurrentRoom().Game.HandlePacket(session.CurrentGamePlayerID(), packet)
			session.EnqueuePlayerPauseState()
			return true
		}
		return false
	}
	if session.CurrentRoom() == nil || session.CurrentGamePlayerID() == "" {
		return true
	}

	session.CurrentRoom().Game.HandlePacket(session.CurrentGamePlayerID(), packet)
	return true
}
