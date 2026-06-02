package game

func (game *Game) HandlePacket(playerID string, packet ClientPacket) {
	game.mu.Lock()
	defer game.mu.Unlock()

	if packet.Type == PacketTypeRespawn {
		game.respawnPlayer(playerID)
		return
	}
	if packet.Type == PacketTypeClientConfig {
		if packet.Config.VisibleWorldWidth > 0 && packet.Config.VisibleWorldHeight > 0 {
			if session, ok := game.playerSessions[playerID]; ok && session != nil {
				session.Config = packet.Config
			}
			if cameraView, ok := game.cameraViews[playerID]; ok && cameraView != nil {
				cameraView.SetConfig(packet.Config)
			}
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
		if packet.Config.VisibleWorldWidth > 0 && packet.Config.VisibleWorldHeight > 0 {
			player.SetConfig(packet.Config)
		}
	}
}
