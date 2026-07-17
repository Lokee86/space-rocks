package inbound

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	targeting "github.com/Lokee86/space-rocks/services/game-server/internal/game/targeting"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
)

type gameplaySession interface {
	CurrentSessionContext() SessionContext
	EnqueuePlayerPauseState()
	EnqueueResyncRequest(SessionContext, realtime.ResyncRequest) bool
}

func HandleGameplayPacket(session gameplaySession, packet game.ClientPacket) bool {
	context := session.CurrentSessionContext()
	if packet.Type != game.PacketTypeInput && packet.Type != game.PacketTypeRespawn && packet.Type != game.PacketTypeClientConfig {
		if context.Room == nil || context.GamePlayerID == "" {
			return false
		}
		gameplayContext := context.Room.GameplayContext()
		if gameplayContext.Game == nil {
			return false
		}
		switch packet.Type {
		case game.PacketTypeResyncRequest:
			matchID := gameplayContext.MatchID
			if matchID == "" || packet.MatchID == "" || packet.MatchID != matchID {
				return false
			}
			request := realtime.ResyncRequest{MatchID: packet.MatchID, Lane: realtime.Lane(packet.Lane), BaselineID: packet.BaselineID, Sequence: packet.Sequence, Reason: packet.Reason}
			if !realtime.IsBaselineLane(request.Lane) {
				return false
			}
			return session.EnqueueResyncRequest(context, request)
		case game.PacketTypeSetTargetPlayerRequest:
			gameplayContext.Game.SetPlayerTarget(context.GamePlayerID, packet.TargetID)
			return true
		case game.PacketTypeSelectTargetAtPositionRequest:
			gameplayContext.Game.SelectTargetAtPosition(context.GamePlayerID, packet.X, packet.Y, targeting.TargetRef{Kind: targeting.TargetKind(packet.TargetKind), ID: packet.TargetID})
			return true
		case game.PacketTypeClearTargetRequest:
			gameplayContext.Game.ClearTarget(context.GamePlayerID)
			return true
		case game.PacketTypePauseRequest:
			gameplayContext.Game.HandlePacket(context.GamePlayerID, packet)
			session.EnqueuePlayerPauseState()
			return true
		}
		return false
	}
	if context.Room == nil || context.GamePlayerID == "" {
		return true
	}
	gameplayContext := context.Room.GameplayContext()
	if gameplayContext.Game == nil {
		return true
	}
	gamePlayerID := context.GamePlayerID
	if packet.Type == game.PacketTypeRespawn {
		gameplayContext.Game.HandlePacket(gamePlayerID, packet)
		return true
	}
	gameplayContext.Game.HandlePacket(gamePlayerID, packet)
	return true
}
