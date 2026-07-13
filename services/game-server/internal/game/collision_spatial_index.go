package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

const defaultSpatialCellSize = 256.0

func (game *Game) rebuildAsteroidSpatialIndex() {
	game.spatialEntries = game.spatialEntries[:0]
	for id, asteroid := range game.entities.Asteroids {
		if asteroid == nil || asteroid.IsPendingDespawn() {
			continue
		}
		body, ok := asteroid.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		game.spatialEntries = append(game.spatialEntries, spatial.Entry{
			Ref:      spatial.Ref{Kind: spatial.KindAsteroid, ID: id},
			Position: body.Position,
			Radius:   physics.BoundingRadius(body.Shape),
		})
	}
	game.spatialIndex.Rebuild(game.spatialEntries)
}

func (game *Game) rebuildPickupSpatialIndex() {
	game.spatialEntries = game.spatialEntries[:0]
	for id, pickup := range game.entities.Pickups {
		if pickup == nil {
			continue
		}
		body, ok := pickup.CollisionBody(game.collisionShapes)
		if !ok {
			continue
		}
		game.spatialEntries = append(game.spatialEntries, spatial.Entry{
			Ref:      spatial.Ref{Kind: spatial.KindPickup, ID: id},
			Position: body.Position,
			Radius:   physics.BoundingRadius(body.Shape),
		})
	}
	game.spatialIndex.Rebuild(game.spatialEntries)
}

