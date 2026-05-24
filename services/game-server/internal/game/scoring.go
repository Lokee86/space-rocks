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
	if player.Paused || player.IsInvulnerable() {
		return
	}

	player.AddScore(award.Points)
	if session, ok := game.playerSessions[award.PlayerID]; ok {
		session.Score = player.Score
	}
	logging.Game.Debug("score awarded",
		logging.FieldPlayerID, award.PlayerID,
		"source", string(award.Reason),
		"amount", award.Points,
		"score", player.Score,
	)
}
