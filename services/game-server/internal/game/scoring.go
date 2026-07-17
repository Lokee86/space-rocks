package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/scoring"
)

func (game *Game) awardScore(award scoring.Award) {
	if award.Points <= 0 {
		return
	}

	player, ok := game.entities.Players[award.PlayerID]
	if !ok {
		return
	}
	if !game.playerCanReceiveScore(award.PlayerID, player) {
		return
	}

	game.addPlayerScoreLocked(award.PlayerID, award.Points)
}
