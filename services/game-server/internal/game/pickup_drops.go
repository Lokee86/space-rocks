package game

import (
	"math/rand"

	"github.com/Lokee86/space-rocks/server/internal/game/drops"
	"github.com/Lokee86/space-rocks/server/internal/game/events"
	"github.com/Lokee86/space-rocks/server/internal/game/entities/pickups"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func (game *Game) maybeDropPickupFromAsteroidLocked(asteroid *runtime.Asteroid) {
	if asteroid == nil {
		return
	}

	table, ok := game.dropTables.ByID["basicasteroids"]
	if !ok {
		return
	}
	if len(game.entities.Pickups) >= table.MaxActivePickups {
		return
	}

	source := drops.Source{
		Type: drops.SourceTypeAsteroid,
		ID:   asteroid.ID,
		Size: asteroid.Size,
		X:    asteroid.X,
		Y:    asteroid.Y,
	}
	rollValues := make([]float64, len(table.Entries))
	for index := range rollValues {
		rollValues[index] = rand.Float64()
	}
	results := game.dropTables.Roll("basicasteroids", source, drops.Roll{Values: rollValues})
	if len(results) == 0 {
		return
	}

	for _, result := range results {
		if len(game.entities.Pickups) >= table.MaxActivePickups {
			return
		}

		pickup, ok, err := game.spawnPickupLocked(pickups.PickupType(result.PickupType), physics.Vector2{X: result.X, Y: result.Y})
		if err != nil || !ok || pickup == nil {
			return
		}

		game.recordDomainEvent(events.Event{
			Type:       events.EventPickupDropped,
			PickupID:   pickup.ID,
			PickupType: string(pickup.Type),
			SourceType: string(source.Type),
			SourceID:   source.ID,
			TableID:    "basicasteroids",
			X:          pickup.X,
			Y:          pickup.Y,
		})
	}
}
