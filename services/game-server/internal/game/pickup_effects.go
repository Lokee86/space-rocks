package game

import (
	pickuprules "github.com/Lokee86/space-rocks/server/internal/game/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
)

func (game *Game) applyPickupEffectIntentLocked(intent pickuprules.EffectIntent) bool {
	if intent.EffectType == "" {
		return false
	}

	switch intent.EffectType {
	case pickuprules.EffectTypeAddLives:
		change := game.addPlayerLivesLocked(intent.PlayerID, intent.Amount)
		if !change.Found {
			return false
		}
		game.recordDomainEvent(events.Event{
			Type:       events.EventPickupEffectApplied,
			PlayerID:   intent.PlayerID,
			PickupID:   intent.PickupID,
			PickupType: intent.PickupType,
			EffectType: intent.EffectType,
			Amount:     intent.Amount,
			LivesAfter: change.After,
		})
		return true
	default:
		return false
	}
}
