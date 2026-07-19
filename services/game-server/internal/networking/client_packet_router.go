package networking

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game"
	"github.com/Lokee86/space-rocks/services/game-server/internal/logging"
	"github.com/Lokee86/space-rocks/services/game-server/internal/networking/inbound"
	"github.com/Lokee86/space-rocks/services/game-server/internal/protocol/packetcodec"
	observability "github.com/Lokee86/space-rocks/shared/go/observabilityevent"
)

func handleClientPacket(session *webSocketSession, remoteAddr string, msg []byte, envelope inbound.ClientPacketEnvelope) {
	adapter := newInboundSessionAdapter(session)
	inbound.RouteClientPacket(inbound.ClientPacketRouter{
		HandleWebRTCSignaling: func() bool {
			return inbound.HandleWebRTCSignalingPacket(adapter, remoteAddr, msg, envelope)
		},
		DecodePacket: func() (game.ClientPacket, error) {
			var packet game.ClientPacket
			if err := packetcodec.Decode(msg, &packet); err != nil {
				context := adapter.CurrentSessionContext()
				logging.Emit(observability.Request{
					Event: observability.EventNamePacketDecodeFailed,
					Context: observability.Context{
						TraceID:    adapter.ConnectionTraceID(),
						SessionID:  adapter.SessionID(),
						RoomID:     context.RoomID,
						PlayerID:   context.GamePlayerID,
						PacketType: envelope.Type,
					},
					Fields: observability.Fields{
						"error_code":   "packet_decode_failed",
						"failure_mode": "packet_decode_failed",
					},
				})
				return game.ClientPacket{}, err
			}
			return packet, nil
		},
		HandleAuth: func(packet game.ClientPacket) bool {
			return inbound.HandleAuthPacket(adapter, packet)
		},
		HandleLobby: func(packet game.ClientPacket) bool {
			return inbound.HandleLobbyPacket(adapter, packet)
		},
		HandleGameplay: func(packet game.ClientPacket) bool {
			return inbound.HandleGameplayPacket(adapter, packet)
		},
	})
}
