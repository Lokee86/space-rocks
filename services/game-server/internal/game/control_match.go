package game

import "github.com/Lokee86/space-rocks/services/game-server/internal/game/rules"

func (target *Control) MatchDecision() rules.MatchDecision {
	return target.game.MatchDecision()
}
