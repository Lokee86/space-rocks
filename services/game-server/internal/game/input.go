package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"

func (game *Game) HandlePacket(playerID string, packet ClientPacket) {
	if packet.Type == PacketTypeInput {
		game.enqueuePlayerInput(playerID, packet.Input)
		return
	}

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

	player, ok := game.entities.Players[playerID]
	if !ok {
		return
	}
	switch packet.Type {
	case PacketTypePauseRequest:
		game.togglePlayerPaused(playerID)
	case PacketTypeClientConfig:
		if packet.Config.VisibleWorldWidth > 0 && packet.Config.VisibleWorldHeight > 0 {
			player.SetConfig(packet.Config)
		}
	}
}

func (game *Game) enqueuePlayerInput(playerID string, input runtime.InputState) {
	if playerID == "" {
		return
	}
	game.inputMu.Lock()
	game.pendingPlayerInputs[playerID] = input
	game.inputMu.Unlock()
}

func (game *Game) applyPendingPlayerInputsLocked() {
	game.inputMu.Lock()
	pending := game.pendingPlayerInputs
	game.pendingPlayerInputs = make(map[string]runtime.InputState, len(pending))
	game.inputMu.Unlock()

	for playerID, input := range pending {
		player, ok := game.entities.Players[playerID]
		if !ok || player == nil || !game.playerCanReceiveInput(playerID, player) {
			continue
		}
		player.SetInput(input)
	}
}
