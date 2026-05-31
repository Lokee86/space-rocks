package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/scoring"
	"github.com/Lokee86/space-rocks/server/internal/logging"
)

func (game *Game) awardScore(award scoring.Award) {
	if award.Points <= 0 {
		return
	}

	player, ok := game.state.Players[award.PlayerID]
	if !ok {
		return
	}
	if !game.playerCanReceiveScore(award.PlayerID, player) {
		return
	}

	change := game.addPlayerScoreLocked(award.PlayerID, award.Points)
	logging.Game.Debug("score awarded",
		logging.FieldPlayerID, award.PlayerID,
		"source", string(award.Reason),
		"amount", award.Points,
		"score", change.After,
	)
}
