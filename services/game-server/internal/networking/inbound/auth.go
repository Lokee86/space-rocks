package inbound

import "github.com/Lokee86/space-rocks/services/game-server/internal/game"

type authSession interface {
	HandleAuthenticateRequest(token string, traceID string)
}

func HandleAuthPacket(session authSession, packet game.ClientPacket) bool {
	if packet.Type != game.PacketTypeAuthenticateRequest {
		return false
	}

	session.HandleAuthenticateRequest(packet.Token, packet.TraceID)
	return true
}
