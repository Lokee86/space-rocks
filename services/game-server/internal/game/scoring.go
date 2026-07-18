package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/scoring"
)

// awardScore preserves the legacy scoring adapter while routing mutation
// through the authoritative gameplay-awards runtime.
func (game *Game) awardScore(award scoring.Award) {
	if award.Points <= 0 || !game.playerCanReceiveDestructionAwardLocked(award.PlayerID) {
		return
	}
	_, _ = game.applyAwardMutationsLocked(game.nextAwardEventIDLocked("legacy_score"), []awards.Mutation{{
		Owner:     awards.Owner{Scope: awards.ScopePlayer, ID: award.PlayerID},
		CounterID: awards.CounterScore, Operation: awards.MutationIncrement,
		Value: float64(award.Points), Source: string(award.Reason),
	}})
}
