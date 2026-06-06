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
	spec := bullet.DamageSpec
	if spec.Amount == 0 {
		spec = damage.DamageSpec{
			Amount: bullet.Damage,
			Type:   damage.DamageTypeKinetic,
			Cause:  damage.DamageCauseProjectile,
		}
	}

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
			Modifiers:  asteroid.DamageModifiers,
		},
		Spec: spec,
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
			Modifiers:  player.DamageModifiers,
		},
		Spec: damage.DamageSpec{
			Amount: asteroid.CollisionDamage,
			Type:   damage.DamageTypeKinetic,
			Cause:  damage.DamageCauseCollision,
		},
	}
}

