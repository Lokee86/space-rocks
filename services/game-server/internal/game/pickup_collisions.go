package game

import pickuprules "github.com/Lokee86/space-rocks/server/internal/game/pickups"
import "github.com/Lokee86/space-rocks/server/internal/game/events"

func (game *Game) handlePlayerPickupCollisions() {
	for playerID, player := range game.entities.Players {
		if player.IsPendingDespawn() {
			continue
		}

		for _, pickup := range game.entities.Pickups {
			if pickup == nil {
				continue
			}

			_, ok := detectPlayerPickupCollision(playerID, player, pickup, game.collisionShapes)
			if !ok {
				continue
			}

			game.removePickupLocked(pickup.ID)

			collection := pickuprules.ResolveCollection(pickuprules.CollectionRequest{
				PlayerID:   playerID,
				PickupID:   pickup.ID,
				PickupType: string(pickup.Type),
				X:          pickup.X,
				Y:          pickup.Y,
			})
			if collection.Collected {
				game.recordDomainEvent(events.Event{
					Type:       events.EventPickupCollected,
					PlayerID:   collection.PlayerID,
					PickupID:   collection.PickupID,
					PickupType: collection.PickupType,
					X:          collection.X,
					Y:          collection.Y,
				})
				game.applyPickupEffectIntentLocked(collection.EffectIntent)
			}
			break
		}
	}
}
