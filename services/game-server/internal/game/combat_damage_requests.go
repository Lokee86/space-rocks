package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func projectileAsteroidDamageRequest(
	collision ProjectileAsteroidCollision,
	bullet *runtime.Bullet,
	asteroid *runtime.Asteroid,
) damage.DamageResolutionRequest {
	return damage.DamageResolutionRequest{
		Source: damage.DamageSource{
			EntityID:   collision.ProjectileID,
			EntityType: damage.EntityTypeProjectile,
			Cause:      damage.DamageCauseProjectile,
		},
		Target: damage.DamageTarget{
			EntityID:   collision.AsteroidID,
			EntityType: damage.EntityTypeAsteroid,
			Health:     asteroid.Health,
		},
		Spec: damage.DamageSpec{
			Amount: bullet.Damage,
			Kind:   damage.DamageKindKinetic,
			Cause:  damage.DamageCauseProjectile,
		},
	}
}

func playerAsteroidDamageRequest(
	collision PlayerAsteroidCollision,
	asteroidID string,
	player *runtime.Ship,
	asteroid *runtime.Asteroid,
) damage.DamageResolutionRequest {
	return damage.DamageResolutionRequest{
		Source: damage.DamageSource{
			EntityID:   asteroidID,
			EntityType: damage.EntityTypeAsteroid,
			Cause:      damage.DamageCauseCollision,
		},
		Target: damage.DamageTarget{
			EntityID:   collision.PlayerID,
			EntityType: damage.EntityTypePlayer,
			Health:     player.Health,
			Shield:     player.Shields,
		},
		Spec: damage.DamageSpec{
			Amount: asteroid.CollisionDamage,
			Kind:   damage.DamageKindKinetic,
			Cause:  damage.DamageCauseCollision,
		},
	}
}
