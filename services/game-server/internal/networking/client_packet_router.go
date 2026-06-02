package networking

import (
	"github.com/Lokee86/space-rocks/server/internal/game"
	"github.com/Lokee86/space-rocks/server/internal/logging"
	"github.com/Lokee86/space-rocks/server/internal/protocol/packetcodec"
	"github.com/Lokee86/space-rocks/server/internal/networking/inbound"
)

func handleClientPacket(session *webSocketSession, remoteAddr string, msg []byte, envelope inbound.ClientPacketEnvelope) {
	adapter := newInboundSessionAdapter(session)
	inbound.RouteClientPacket(inbound.ClientPacketRouter{
		HandleSimpleDevtools: func() bool {
			return handleSimpleDevtoolsPacket(session, remoteAddr, msg, envelope)
		},
		HandlePlacementDevtools: func() bool {
			return handlePlacementDevtoolsPacket(session, remoteAddr, msg, envelope)
		},
		HandleRemainingDevtools: func() bool {
			return handleRemainingDevtoolsPacket(session, remoteAddr, msg, envelope)
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
		HandleTelemetry: func(packet game.ClientPacket) bool {
			return inbound.HandleTelemetryPacket(adapter, remoteAddr, packet)
		},
		HandleLobby: func(packet game.ClientPacket) bool {
			if inbound.HandleLobbyPacket(adapter, packet) {
				return true
			}
			return handleLobbyPacket(session, packet)
		},
		HandleGameplay: func(packet game.ClientPacket) bool {
			return handleGameplayPacket(session, packet)
		},
	})
}
