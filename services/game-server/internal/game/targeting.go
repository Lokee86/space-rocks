package game

import targetpolicy "github.com/Lokee86/space-rocks/server/internal/game/targeting"

func (game *Game) SetPlayerTarget(playerID string, targetPlayerID string) bool {
	game.mu.Lock()
	defer game.mu.Unlock()

	acceptedTargetPlayerID, ok := targetpolicy.ValidateRequestedTarget(
		playerID,
		targetPlayerID,
		game.playerExistsLocked,
	)
	if !ok {
		return false
	}

	player, exists := game.state.Players[playerID]
	if !exists || player == nil {
		return false
	}

	player.TargetPlayerID = acceptedTargetPlayerID
	return true
}

func (game *Game) ClearPlayerTarget(playerID string) bool {
	return game.SetPlayerTarget(playerID, "")
}

func (game *Game) PlayerTarget(playerID string) string {
	game.mu.Lock()
	defer game.mu.Unlock()

	player, exists := game.state.Players[playerID]
	if !exists || player == nil {
		return ""
	}

	return player.TargetPlayerID
}

func (game *Game) playerExistsLocked(playerID string) bool {
	player, exists := game.state.Players[playerID]
	return exists && player != nil
}

func (game *Game) clearTargetsForMissingPlayersLocked() {
	for _, player := range game.state.Players {
		if player == nil {
			continue
		}
		targetPlayerID := player.TargetPlayerID
		if targetPlayerID == "" {
			continue
		}
		if _, exists := game.playerSessions[targetPlayerID]; exists {
			continue
		}
		player.TargetPlayerID = ""
	}
}
