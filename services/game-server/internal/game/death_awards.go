package game

import (
	"fmt"
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/awards"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/lives"
)

func (game *Game) applyDeathAwardsLocked(fact lives.DeathFact) {
	runtime := game.awardsRuntime()
	if !fact.Accepted || fact.PlayerID == "" {
		return
	}
	eventID := fmt.Sprintf("death:%s:%d", fact.PlayerID, fact.DeathCount)
	victimOwner := awards.Owner{Scope: awards.ScopePlayer, ID: fact.PlayerID}
	if runtime.EventProcessed(eventID) {
		runtime.ClearContributions(fact.PlayerID)
		return
	}

	mutations := []awards.Mutation{{
		Owner: victimOwner, CounterID: awards.CounterDeaths, Operation: awards.MutationIncrement,
		Value: 1, Source: fact.ReasonCode,
	}}
	killerID := ""
	if fact.Input.Attribution == lives.AttributionPlayerCaused &&
		fact.Input.KillerPlayerID != "" && fact.Input.KillerPlayerID != fact.PlayerID &&
		game.playerCanReceiveAwardsLocked(fact.Input.KillerPlayerID) {
		killerID = fact.Input.KillerPlayerID
		mutations = append(mutations, awards.Mutation{
			Owner:     awards.Owner{Scope: awards.ScopePlayer, ID: killerID},
			CounterID: awards.CounterKills, Operation: awards.MutationIncrement,
			Value: 1, Source: fact.ReasonCode,
		})
	}

	assistIDs := make(map[string]struct{})
	if game.awardPolicy.Assists.Enabled {
		for _, credit := range runtime.ResolveAssists(
			fact.PlayerID, killerID, game.awardClock, game.awardPolicy.Assists, game.eligibleAwardPlayersLocked(),
		) {
			assistIDs[credit.PlayerID] = struct{}{}
		}
		for _, playerID := range fact.Input.AssistPlayerIDs {
			if playerID != "" && playerID != killerID && game.playerCanReceiveAwardsLocked(playerID) {
				assistIDs[playerID] = struct{}{}
			}
		}
	}
	orderedAssistIDs := make([]string, 0, len(assistIDs))
	for playerID := range assistIDs {
		orderedAssistIDs = append(orderedAssistIDs, playerID)
	}
	sort.Strings(orderedAssistIDs)
	for _, playerID := range orderedAssistIDs {
		mutations = append(mutations, awards.Mutation{
			Owner:     awards.Owner{Scope: awards.ScopePlayer, ID: playerID},
			CounterID: awards.CounterAssists, Operation: awards.MutationIncrement,
			Value: 1, Source: "assist",
		})
		if game.awardPolicy.AssistScore > 0 {
			mutations = append(mutations, awards.Mutation{
				Owner:     awards.Owner{Scope: awards.ScopePlayer, ID: playerID},
				CounterID: awards.CounterScore, Operation: awards.MutationIncrement,
				Value: game.awardPolicy.AssistScore, Source: "assist",
			})
		}
	}

	result, err := game.applyAwardMutationsLocked(eventID, mutations)
	if err == nil && result.Applied && killerID != "" {
		_, _ = runtime.IncrementStreak(
			awards.Owner{Scope: awards.ScopePlayer, ID: killerID}, game.awardPolicy.KillStreakName,
		)
	}
	runtime.ResetProgress(victimOwner)
	runtime.ClearContributions(fact.PlayerID)
}
