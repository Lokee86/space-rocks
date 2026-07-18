package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/player"

func (game *Game) playerWorldStateLocked(playerID string) (player.WorldState, bool) {
	session, ok := game.playerSessions[playerID]
	if !ok || session == nil {
		return player.WorldState{}, false
	}
	lifecycle, ok := game.lifeRuntime.ParticipantSnapshot(playerID)
	if !ok {
		return player.WorldState{}, false
	}

	positionX := session.SpawnPosition.X
	positionY := session.SpawnPosition.Y
	hasActiveShip := false

	ship, shipExists := game.entities.Players[playerID]
	if shipExists && ship != nil && !ship.IsPendingDespawn() {
		hasActiveShip = true
		position := ship.Position()
		positionX = position.X
		positionY = position.Y
	} else if cameraView, hasCameraView := game.cameraViews[playerID]; hasCameraView && cameraView != nil {
		positionX = cameraView.X
		positionY = cameraView.Y
	}

	return player.BuildWorldState(player.BuildWorldStateInput{
		ID:              playerID,
		Status:          lifecycle.Status,
		HasActiveShip:   hasActiveShip,
		X:               positionX,
		Y:               positionY,
		Lives:           game.projectedPlayerLives(playerID, lifecycle),
		RespawnCooldown: lifecycle.RespawnCooldown,
	}), true
}

func (game *Game) playerWorldStatesLocked() map[string]player.WorldState {
	playerStates := make(map[string]player.WorldState, len(game.playerSessions))
	for playerID := range game.playerSessions {
		state, ok := game.playerWorldStateLocked(playerID)
		if !ok {
			continue
		}
		playerStates[playerID] = state
	}
	return playerStates
}
