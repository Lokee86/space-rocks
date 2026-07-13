package game

import (
	"sort"

	"github.com/Lokee86/space-rocks/services/game-server/internal/game/physics"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/space"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/spatial"
)

func (game *Game) collisionPlayerIDsSorted() []string {
	game.collisionPlayerIDs = game.collisionPlayerIDs[:0]
	for id, player := range game.entities.Players {
		if player != nil {
			game.collisionPlayerIDs = append(game.collisionPlayerIDs, id)
		}
	}
	sort.Strings(game.collisionPlayerIDs)
	return game.collisionPlayerIDs
}

func (game *Game) collisionProjectileIDsSorted() []string {
	game.collisionProjectileIDs = game.collisionProjectileIDs[:0]
	for id, projectile := range game.entities.Projectiles {
		if projectile != nil {
			game.collisionProjectileIDs = append(game.collisionProjectileIDs, id)
		}
	}
	sort.Strings(game.collisionProjectileIDs)
	return game.collisionProjectileIDs
}

func (game *Game) asteroidCollisionCandidates(body physics.CollisionBody) []spatial.Ref {
	game.spatialRefs = game.spatialRefs[:0]
	game.spatialRefs = game.spatialIndex.QueryCircle(game.spatialRefs, body.Position, physics.BoundingRadius(body.Shape), spatial.KindMask(spatial.KindAsteroid))
	game.compactAsteroidRefs()
	game.sortAsteroidRefsByDistance(body.Position)
	return game.spatialRefs
}

func (game *Game) pickupCollisionCandidates(body physics.CollisionBody) []spatial.Ref {
	game.spatialRefs = game.spatialRefs[:0]
	game.spatialRefs = game.spatialIndex.QueryCircle(game.spatialRefs, body.Position, physics.BoundingRadius(body.Shape), spatial.KindMask(spatial.KindPickup))
	game.compactPickupRefs()
	game.sortPickupRefsByDistance(body.Position)
	return game.spatialRefs
}

func (game *Game) compactAsteroidRefs() {
	refs := game.spatialRefs[:0]
	for _, ref := range game.spatialRefs {
		asteroid := game.entities.Asteroids[ref.ID]
		if asteroid != nil && !asteroid.IsPendingDespawn() {
			refs = append(refs, ref)
		}
	}
	game.spatialRefs = refs
}

func (game *Game) compactPickupRefs() {
	refs := game.spatialRefs[:0]
	for _, ref := range game.spatialRefs {
		if game.entities.Pickups[ref.ID] != nil {
			refs = append(refs, ref)
		}
	}
	game.spatialRefs = refs
}

func (game *Game) sortAsteroidRefsByDistance(center physics.Vector2) {
	sort.Slice(game.spatialRefs, func(i, j int) bool {
		left := game.entities.Asteroids[game.spatialRefs[i].ID].Position()
		right := game.entities.Asteroids[game.spatialRefs[j].ID].Position()
		leftDistance := space.ShortestDelta(center, left, space.DefaultBounds()).LengthSquared()
		rightDistance := space.ShortestDelta(center, right, space.DefaultBounds()).LengthSquared()
		if leftDistance == rightDistance {
			return game.spatialRefs[i].ID < game.spatialRefs[j].ID
		}
		return leftDistance < rightDistance
	})
}

func (game *Game) sortPickupRefsByDistance(center physics.Vector2) {
	sort.Slice(game.spatialRefs, func(i, j int) bool {
		left := game.entities.Pickups[game.spatialRefs[i].ID].Position()
		right := game.entities.Pickups[game.spatialRefs[j].ID].Position()
		leftDistance := space.ShortestDelta(center, left, space.DefaultBounds()).LengthSquared()
		rightDistance := space.ShortestDelta(center, right, space.DefaultBounds()).LengthSquared()
		if leftDistance == rightDistance {
			return game.spatialRefs[i].ID < game.spatialRefs[j].ID
		}
		return leftDistance < rightDistance
	})
}