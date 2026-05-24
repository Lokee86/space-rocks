package game

import "github.com/Lokee86/space-rocks/server/internal/game/rules"

func (game *Game) matchSnapshot() rules.MatchSnapshot {
	players := make([]rules.PlayerSnapshot, 0, len(game.playerSessions))
	for playerID, session := range game.playerSessions {
		_, hasActiveShip := game.state.Players[playerID]
		players = append(players, rules.PlayerSnapshot{
			ID:                session.ID,
			HasActiveShip:     hasActiveShip,
			HasRemainingLives: session.Lives > 0,
		})
	}
	return rules.MatchSnapshot{Players: players}
}
