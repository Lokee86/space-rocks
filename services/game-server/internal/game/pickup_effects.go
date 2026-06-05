package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
)

func (game *Game) applyPickupEffect(playerID string, pickup *pickups.Pickup) bool {
	if pickup == nil {
		return false
	}

	switch pickup.Type {
	case pickups.TypeOneUp:
		change := game.addPlayerLivesLocked(playerID, 1)
		if !change.Found {
			return false
		}
		game.recordDomainEvent(events.Event{
			Type:       events.EventPickupCollected,
			PlayerID:   playerID,
			PickupID:   pickup.ID,
			PickupType: string(pickup.Type),
			LivesAfter: change.After,
		})
		return true
	default:
		return false
	}
}
