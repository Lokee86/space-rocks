package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
)

func (game *Game) recordDamageAwardConsequences(request damage.DamageResolutionRequest, result damage.DamageResult) {
	applied := result.AppliedToHealth + result.AbsorbedByShield
	if applied <= 0 || result.Kind != damage.DamageResultKindDamage {
		return
	}

	mutations := make([]awards.Mutation, 0, 2)
	sourcePlayerID := request.Source.ResponsiblePlayerID
	if sourcePlayerID != "" && game.playerCanReceiveAwardsLocked(sourcePlayerID) {
		mutations = append(mutations, awards.Mutation{
			Owner:     awards.Owner{Scope: awards.ScopePlayer, ID: sourcePlayerID},
			CounterID: awards.CounterDamageDealt, Operation: awards.MutationIncrement,
			Value: float64(applied), Source: "damage_applied",
		})
	}
	if request.Target.EntityType == damage.EntityTypePlayer && game.playerCanReceiveAwardsLocked(request.Target.EntityID) {
		mutations = append(mutations, awards.Mutation{
			Owner:     awards.Owner{Scope: awards.ScopePlayer, ID: request.Target.EntityID},
			CounterID: awards.CounterDamageTaken, Operation: awards.MutationIncrement,
			Value: float64(applied), Source: "damage_applied",
		})
	}
	if len(mutations) > 0 {
		_, _ = game.applyAwardMutationsLocked(game.nextAwardEventIDLocked("damage"), mutations)
	}

	if sourcePlayerID == "" || !game.playerCanReceiveAwardsLocked(sourcePlayerID) {
		return
	}
	teamID := request.Source.ResponsibleTeamID
	if teamID == "" {
		if resolved, ok := game.playerTeamLocked(sourcePlayerID); ok {
			teamID = string(resolved)
		}
	}
	_ = game.awardsRuntime().RecordContribution(awards.Contribution{
		TargetID: request.Target.EntityID, PlayerID: sourcePlayerID, TeamID: teamID,
		Amount: float64(applied), At: game.awardClock,
	})
}
