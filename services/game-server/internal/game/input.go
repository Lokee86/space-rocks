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
	if game.handleDebugPacket(playerID, player, packet) {
		return
	}
	switch packet.Type {
	case PacketTypeInput:
		if !player.CanReceiveInput() {
			return
		}
		player.SetInput(packet.Input)
	case PacketTypePausePlayer:
		game.setPlayerPaused(playerID, true)
	case PacketTypeResumePlayer:
		game.setPlayerPaused(playerID, false)
	case PacketTypePauseRequest:
		game.togglePlayerPaused(playerID)
	case PacketTypeClientConfig:
		player.SetConfig(packet.Config)
	}
}
