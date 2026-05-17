package game

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/entities"
)

type ScoreSource int

const (
	ScoreSourceAsteroid ScoreSource = iota
)

type ScoreAward struct {
	PlayerID string
	Source   ScoreSource
	Amount   int
}

func NewAsteroidHitScoreAward(playerID string, asteroid *entities.Asteroid) ScoreAward {
	amount := 0
	if asteroid.Size > 0 {
		amount = constants.BaseScore / asteroid.Size
	}

	return ScoreAward{
		PlayerID: playerID,
		Source:   ScoreSourceAsteroid,
		Amount:   amount,
	}
}

func (game *Game) awardScore(award ScoreAward) {
	if award.Amount <= 0 {
		return
	}

	player, ok := game.state.Players[award.PlayerID]
	if !ok {
		return
	}

	player.AddScore(award.Amount)
	if session, ok := game.playerSessions[award.PlayerID]; ok {
		session.Score = player.Score
	}
}
