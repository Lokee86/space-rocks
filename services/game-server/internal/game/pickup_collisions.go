package game

func (game *Game) handlePlayerPickupCollisions() {
	for playerID, player := range game.entities.Players {
		if player.IsPendingDespawn() {
			continue
		}

		for pickupID, pickup := range game.entities.Pickups {
			if pickup == nil {
				continue
			}

			_, ok := detectPlayerPickupCollision(playerID, player, pickup, game.collisionShapes)
			if !ok {
				continue
			}

			game.applyPickupEffect(playerID, pickup)
			game.removePickupLocked(pickupID)
			break
		}
	}
}
