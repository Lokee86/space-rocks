package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/events"
	pickuprules "github.com/Lokee86/space-rocks/services/game-server/internal/game/pickups"
)

func (game *Game) handlePlayerPickupCollisions() {
	for _, playerID := range game.collisionPlayerIDsSorted() {
		player := game.entities.Players[playerID]
		if player == nil {
			continue
		}
		if player.IsPendingDespawn() {
			continue
		}

		playerBody, ok := player.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}

		for _, pickupRef := range game.pickupCollisionCandidates(playerBody) {
			pickup := game.entities.Pickups[pickupRef.ID]
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
