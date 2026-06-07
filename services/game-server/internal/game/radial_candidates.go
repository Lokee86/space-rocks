package game

import (
	"math"

	"github.com/Lokee86/space-rocks/server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

func (game *Game) radialCandidates() []radial.Candidate {
	candidates := make([]radial.Candidate, 0,
		len(game.entities.Asteroids)+
			len(game.entities.Enemies)+
			len(game.entities.Players)+
			len(game.entities.Projectiles)+
			len(game.entities.Pickups),
	)

	for id, asteroid := range game.entities.Asteroids {
		if asteroid == nil || asteroid.IsPendingDespawn() {
			continue
		}

		candidates = append(candidates, radial.Candidate{
			ID:       id,
			Kind:     radial.TargetAsteroid,
			Position: asteroid.Position(),
			Radius:   asteroidRadialCandidateRadius(asteroid.CollisionBody(game.collisionShapes)),
		})
	}

	for id, enemy := range game.entities.Enemies {
		if enemy == nil || enemy.IsPendingDespawn() {
			continue
		}

		candidates = append(candidates, radial.Candidate{
			ID:       id,
			Kind:     radial.TargetEnemy,
			Position: enemy.Position(),
		})
	}

	for id, player := range game.entities.Players {
		if player == nil || player.IsPendingDespawn() {
			continue
		}

		candidates = append(candidates, radial.Candidate{
			ID:       id,
			Kind:     radial.TargetPlayer,
			Position: player.Position(),
		})
	}

	for id, projectile := range game.entities.Projectiles {
		if projectile == nil || projectile.IsPendingDespawn() {
			continue
		}

		candidates = append(candidates, radial.Candidate{
			ID:       id,
			Kind:     radial.TargetProjectile,
			Position: projectile.Position(),
		})
	}

	for id, pickup := range game.entities.Pickups {
		if pickup == nil {
			continue
		}

		candidates = append(candidates, radial.Candidate{
			ID:       id,
			Kind:     radial.TargetPickup,
			Position: pickup.Position(),
		})
	}

	return candidates
}

func asteroidRadialCandidateRadius(body physics.CollisionBody, ok bool) float64 {
	if !ok {
		return 0
	}
	if body.Shape.Radius > 0 {
		return body.Shape.Radius
	}

	radius := 0.0
	for _, point := range physics.CollisionBodyOutlinePoints(body) {
		distance := point.Subtract(body.Position).Length()
		radius = math.Max(radius, distance)
	}

	return radius
}
