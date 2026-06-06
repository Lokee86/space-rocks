package game

import (
	"fmt"

	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func (game *Game) SpawnPickup(pickupType pickups.PickupType, position physics.Vector2) (*pickups.Pickup, bool, error) {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.spawnPickupLocked(pickupType, position)
}

func (game *Game) spawnPickupLocked(pickupType pickups.PickupType, position physics.Vector2) (*pickups.Pickup, bool, error) {
	definition, ok := pickups.DefinitionFor(pickupType)
	if !ok {
		return nil, false, fmt.Errorf("unknown pickup type %q", pickupType)
	}

	game.nextPickupID++
	pickupID := fmt.Sprintf("pickup_%d", game.nextPickupID)
	pickup := &pickups.Pickup{
		ID:              pickupID,
		Type:            definition.Type,
		X:               position.X,
		Y:               position.Y,
		Health:          definition.Health,
		AgeSeconds:      0,
		LifespanSeconds: definition.LifespanSeconds,
	}
	game.entities.Pickups[pickupID] = pickup

	return pickup, true, nil
}

func (game *Game) RemovePickup(id string) bool {
	game.mu.Lock()
	defer game.mu.Unlock()

	return game.removePickupLocked(id)
}

func (game *Game) removePickupLocked(id string) bool {
	if _, ok := game.entities.Pickups[id]; !ok {
		return false
	}

	delete(game.entities.Pickups, id)
	return true
}

func (game *Game) pickupStatesLocked() map[string]runtime.PickupState {
	pickupStates := make(map[string]runtime.PickupState, len(game.entities.Pickups))
	for id, pickup := range game.entities.Pickups {
		pickupStates[id] = runtime.PickupState{
			ID:              pickup.ID,
			Type:            string(pickup.Type),
			X:               pickup.X,
			Y:               pickup.Y,
			Health:          pickup.Health,
			AgeSeconds:      pickup.AgeSeconds,
			LifespanSeconds: pickup.LifespanSeconds,
		}
	}

	return pickupStates
}
