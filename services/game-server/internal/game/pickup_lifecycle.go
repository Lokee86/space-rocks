package game

import "github.com/Lokee86/space-rocks/server/internal/game/events"

func (game *Game) stepPickups(delta float64) {
	for id, pickup := range game.entities.Pickups {
		pickup.AgeSeconds += delta
		if pickup.LifespanSeconds > 0 && pickup.AgeSeconds >= pickup.LifespanSeconds {
			game.recordDomainEvent(events.Event{
				Type:       events.EventPickupExpired,
				PickupID:   pickup.ID,
				PickupType: string(pickup.Type),
				X:          pickup.X,
				Y:          pickup.Y,
			})
			delete(game.entities.Pickups, id)
		}
	}
}
