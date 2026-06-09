package inbound

import "github.com/Lokee86/space-rocks/server/internal/game"

type authSession interface {
	HandleAuthenticateRequest(token string)
}

func HandleAuthPacket(session authSession, packet game.ClientPacket) bool {
	if packet.Type != game.PacketTypeAuthenticateRequest {
		return false
	}

	session.HandleAuthenticateRequest(packet.Token)
	return true
}
