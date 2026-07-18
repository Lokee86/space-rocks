package game

import (
	playerstate "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func (game *Game) HandlePacket(playerID string, packet ClientPacket) {
	if packet.Type == PacketTypeInput {
		game.enqueuePlayerInput(playerID, packet.Input)
		return
	}

	game.mu.Lock()
	defer game.mu.Unlock()

	if packet.Type == PacketTypeRespawn {
		if respawned := game.respawnPlayer(playerID); respawned {
			game.participationRuntime.RecordAction(playerID)
		}
		return
	}
	if packet.Type == PacketTypeClientConfig {
		if _, isBot := game.botControllers[playerID]; isBot {
			return
		}
		if packet.Config.VisibleWorldWidth <= 0 || packet.Config.VisibleWorldHeight <= 0 {
			return
		}

		config := runtime.ClampCameraConfig(packet.Config)
		if session, ok := game.playerSessions[playerID]; ok && session != nil {
			session.Config = config
		}
		if cameraView, ok := game.cameraViews[playerID]; ok && cameraView != nil {
			cameraView.SetConfig(config)
		}
		if player, ok := game.entities.Players[playerID]; ok && player != nil {
			player.SetConfig(config)
		}
		return
	}

	if _, ok := game.entities.Players[playerID]; !ok {
		return
	}
	if packet.Type == PacketTypePauseRequest {
		if player, ok := game.entities.Players[playerID]; ok && player != nil {
			if status, statusOK := game.lifeRuntime.Status(playerID); statusOK && status == playerstate.StatusActive && !player.IsPendingDespawn() {
				game.participationRuntime.RecordAction(playerID)
			}
		}
		game.togglePlayerPaused(playerID)
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
		game.participationRuntime.RecordAction(playerID)
	}
}
