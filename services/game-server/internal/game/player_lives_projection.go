package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"

func (game *Game) projectedPlayerLives(playerID string, state lives.ParticipantState) int {
	projected, ok := game.lifeRuntime.ProjectedLives(playerID)
	if ok {
		return projected
	}
	return state.EffectiveLives
}
