package game

func (game *Game) HandlePacket(playerID string, packet ClientPacket) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if packet.Type == PacketTypeRespawn {
		game.respawnPlayer(playerID)
		return
	}
	if packet.Type == PacketTypeClientConfig {
		if session, ok := game.playerSessions[playerID]; ok {
			session.Config = packet.Config
		}
		if cameraView, ok := game.cameraViews[playerID]; ok {
			cameraView.SetConfig(packet.Config)
		}
	}

	player, ok := game.state.Players[playerID]
	if !ok {
		return
	}
	switch packet.Type {
	case PacketTypeInput:
		if !game.playerCanReceiveInput(playerID, player) {
			return
		}
		player.SetInput(packet.Input)
	case PacketTypePauseRequest:
		game.togglePlayerPaused(playerID)
	case PacketTypeClientConfig:
		player.SetConfig(packet.Config)
	}
}
