package inbound

import "github.com/Lokee86/space-rocks/server/internal/game"

type ClientPacketRouter struct {
	HandleSimpleDevtools    func() bool
	HandlePlacementDevtools func() bool
	HandleRemainingDevtools func() bool
	DecodePacket            func() (game.ClientPacket, error)
	HandleAuth              func(game.ClientPacket) bool
	HandleTelemetry         func(game.ClientPacket) bool
	HandleLobby             func(game.ClientPacket) bool
	HandleGameplay          func(game.ClientPacket) bool
}

func RouteClientPacket(router ClientPacketRouter) {
	if router.HandleSimpleDevtools != nil && router.HandleSimpleDevtools() {
		return
	}
	if router.HandlePlacementDevtools != nil && router.HandlePlacementDevtools() {
		return
	}
	if router.HandleRemainingDevtools != nil && router.HandleRemainingDevtools() {
		return
	}

	packet, err := router.DecodePacket()
	if err != nil {
		return
	}

	if router.HandleAuth != nil && router.HandleAuth(packet) {
		return
	}
	if router.HandleTelemetry != nil && router.HandleTelemetry(packet) {
		return
	}
	if router.HandleLobby != nil && router.HandleLobby(packet) {
		return
	}
	if router.HandleGameplay != nil && router.HandleGameplay(packet) {
		return
	}
}
