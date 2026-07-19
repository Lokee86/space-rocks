package inbound

import "github.com/Lokee86/space-rocks/services/game-server/internal/game"

type ClientPacketRouter struct {
	HandleWebRTCSignaling func() bool
	DecodePacket          func() (game.ClientPacket, error)
	HandleAuth            func(game.ClientPacket) bool
	HandleLobby           func(game.ClientPacket) bool
	HandleGameplay        func(game.ClientPacket) bool
}

func RouteClientPacket(router ClientPacketRouter) {
	if router.HandleWebRTCSignaling != nil && router.HandleWebRTCSignaling() {
		return
	}

	packet, err := router.DecodePacket()
	if err != nil {
		return
	}

	if router.HandleAuth != nil && router.HandleAuth(packet) {
		return
	}
	if router.HandleLobby != nil && router.HandleLobby(packet) {
		return
	}
	if router.HandleGameplay != nil && router.HandleGameplay(packet) {
		return
	}
}
