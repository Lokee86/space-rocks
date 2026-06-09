package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/networking/inbound"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
)

func handleClientPacket(session *webSocketSession, remoteAddr string, msg []byte, envelope inbound.ClientPacketEnvelope) {
	adapter := newInboundSessionAdapter(session)
	inbound.RouteClientPacket(inbound.ClientPacketRouter{
		HandleSimpleDevtools: func() bool {
			return inbound.HandleSimpleDevtoolsPacket(adapter, remoteAddr, msg, envelope)
		},
		HandlePlacementDevtools: func() bool {
			return inbound.HandlePlacementDevtoolsPacket(adapter, remoteAddr, msg, envelope)
		},
		HandleRemainingDevtools: func() bool {
			return inbound.HandleRemainingDevtoolsPacket(adapter, remoteAddr, msg, envelope)
		},
		DecodePacket: func() (game.ClientPacket, error) {
			var packet game.ClientPacket
			if err := packetcodec.Decode(msg, &packet); err != nil {
				logging.Network.Warn("websocket packet decode failed",
					logging.FieldError, err,
					logging.FieldRoomID, adapter.CurrentRoomID(),
					logging.FieldPlayerID, adapter.CurrentGamePlayerID(),
					"session_id", adapter.SessionID(),
					logging.FieldRemoteAddr, remoteAddr,
				)
				return game.ClientPacket{}, err
			}
			return packet, nil
		},
		HandleAuth: func(packet game.ClientPacket) bool {
			return inbound.HandleAuthPacket(adapter, packet)
		},
		HandleTelemetry: func(packet game.ClientPacket) bool {
			return inbound.HandleTelemetryPacket(adapter, remoteAddr, packet)
		},
		HandleLobby: func(packet game.ClientPacket) bool {
			return inbound.HandleLobbyPacket(adapter, packet)
		},
		HandleGameplay: func(packet game.ClientPacket) bool {
			return inbound.HandleGameplayPacket(adapter, packet)
		},
	})
}
