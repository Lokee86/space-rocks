package game

import (
	"math"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/scoring"
)

func (game *Game) applyDestructionAwardsLocked(targetID string, calculated []scoring.Award) {
	runtime := game.awardsRuntime()
	eventID := "destruction:" + targetID
	if targetID == "" || runtime.EventProcessed(eventID) {
		runtime.ClearContributions(targetID)
		return
	}

	mutations := make([]awards.Mutation, 0, len(calculated)*2)
	comboOwners := make([]awards.Owner, 0, len(calculated))
	killerIDs := make(map[string]struct{})
	primaryKillerID := ""
	for _, award := range calculated {
		if award.PlayerID == "" || award.Points <= 0 || !game.playerCanReceiveDestructionAwardLocked(award.PlayerID) {
			continue
		}
		if primaryKillerID == "" {
			primaryKillerID = award.PlayerID
		}
		owner := awards.Owner{Scope: awards.ScopePlayer, ID: award.PlayerID}
		multiplier := 1.0
		if game.awardPolicy.ComboEnabled && game.awardPolicy.ComboCounter == awards.CounterScore {
			multiplier = runtime.Combo(owner, game.awardClock).Multiplier
			comboOwners = append(comboOwners, owner)
		}
		mutations = append(mutations, awards.Mutation{
			Owner: owner, CounterID: awards.CounterScore, Operation: awards.MutationIncrement,
			Value: math.Round(float64(award.Points) * multiplier), Source: string(award.Reason),
		})
		if _, seen := killerIDs[award.PlayerID]; !seen {
			killerIDs[award.PlayerID] = struct{}{}
			mutations = append(mutations, awards.Mutation{
				Owner: owner, CounterID: awards.CounterKills, Operation: awards.MutationIncrement,
				Value: 1, Source: string(award.Reason),
			})
		}
	}

	assistOwners := make([]awards.Owner, 0)
	for _, credit := range runtime.ResolveAssists(
		targetID, primaryKillerID, game.awardClock, game.awardPolicy.Assists, game.eligibleAwardPlayersLocked(),
	) {
		owner := awards.Owner{Scope: awards.ScopePlayer, ID: credit.PlayerID}
		assistOwners = append(assistOwners, owner)
		mutations = append(mutations, awards.Mutation{
			Owner: owner, CounterID: awards.CounterAssists, Operation: awards.MutationIncrement,
			Value: 1, Source: "assist",
		})
		if game.awardPolicy.AssistScore > 0 {
			mutations = append(mutations, awards.Mutation{
				Owner: owner, CounterID: awards.CounterScore, Operation: awards.MutationIncrement,
				Value: game.awardPolicy.AssistScore, Source: "assist",
			})
		}
	}

	result, err := game.applyAwardMutationsLocked(eventID, mutations)
	if err == nil && result.Applied {
		for _, owner := range comboOwners {
			_, _ = runtime.ApplyCombo(owner, game.awardClock)
		}
		if game.awardPolicy.AssistScore > 0 && game.awardPolicy.ComboEnabled {
			for _, owner := range assistOwners {
				_, _ = runtime.ApplyCombo(owner, game.awardClock)
			}
		}
	}
	runtime.ClearContributions(targetID)
}

func (game *Game) playerCanReceiveDestructionAwardLocked(playerID string) bool {
	if !game.playerCanReceiveAwardsLocked(playerID) {
		return false
	}
	player, ok := game.entities.Players[playerID]
	if !ok || player == nil {
		return true
	}
	return game.playerCanReceiveScore(playerID, player)
}
