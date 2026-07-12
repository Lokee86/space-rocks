package inbound

import (
	"sync"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	targeting "github.com/Lokee86/space-rocks/services/game-server/internal/game/targeting"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/realtime"
	"github.com/Lokee86/space-rocks/services/game-server/internal/rooms"
)

type gameplaySession interface {
	CurrentRoom() *rooms.Room
	CurrentGamePlayerID() string
	EnqueuePlayerPauseState()
	EnqueueResyncRequest(realtime.ResyncRequest) bool
}

var loggedInputPackets sync.Map
var loggedRespawnPackets sync.Map

func HandleGameplayPacket(session gameplaySession, packet game.ClientPacket) bool {
	if packet.Type != game.PacketTypeInput && packet.Type != game.PacketTypeRespawn && packet.Type != game.PacketTypeClientConfig {
		room := session.CurrentRoom()
		if room == nil || session.CurrentGamePlayerID() == "" {
			return false
		}
		gameInstance := room.GameInstance()
		switch packet.Type {
		case game.PacketTypeResyncRequest:
			if gameInstance == nil {
				return false
			}
			matchID := room.CurrentMatchID()
			if matchID == "" || packet.MatchID == "" {
				return false
			}
			request := realtime.ResyncRequest{MatchID: packet.MatchID, Lane: realtime.Lane(packet.Lane), BaselineID: packet.BaselineID, Sequence: packet.Sequence, Reason: packet.Reason}
			if !realtime.IsBaselineLane(request.Lane) {
				return false
			}
			return session.EnqueueResyncRequest(request)
		case game.PacketTypeSetTargetPlayerRequest:
			gameInstance.SetPlayerTarget(session.CurrentGamePlayerID(), packet.TargetID)
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

	gamePlayerID := session.CurrentGamePlayerID()
	if packet.Type == game.PacketTypeRespawn {
		if _, loaded := loggedRespawnPackets.LoadOrStore(gamePlayerID, true); !loaded {
			logRespawnPacketReceived(gamePlayerID, packet)
		}
		room.GameInstance().HandlePacket(gamePlayerID, packet)
		return true
	}

	if _, loaded := loggedInputPackets.LoadOrStore(gamePlayerID, true); !loaded {
		logging.Network.Info("gameplay input packet received",
			logging.FieldPlayerID, gamePlayerID,
			"packet_type", packet.Type,
			"forward", packet.Input.Forward,
			"back", packet.Input.Back,
			"left", packet.Input.Left,
			"right", packet.Input.Right,
			"primary_fire", packet.Input.PrimaryFire,
			"secondary_fire", packet.Input.SecondaryFire,
		)
	}

	room.GameInstance().HandlePacket(gamePlayerID, packet)
	return true
}

func logRespawnPacketReceived(gamePlayerID string, packet game.ClientPacket) {
	logging.Network.Info("gameplay respawn packet received",
		logging.FieldPlayerID, gamePlayerID,
		"packet_type", packet.Type,
	)
}
