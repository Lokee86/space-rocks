package game

import (
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/server/internal/game/runtime"
)

func radialDamageRequestFromHitAndAsteroid(hit radial.Hit, asteroid *runtime.Asteroid) damage.DamageResolutionRequest {
	return damage.DamageResolutionRequest{
		Source: damage.DamageSource{
			EntityID:   hit.SourceID,
			EntityType: damage.EntityTypeProjectile,
			Cause:      damage.DamageCauseArea,
		},
		Target: damage.DamageTarget{
			EntityID:   asteroid.ID,
			EntityType: damage.EntityTypeAsteroid,
			Health:     asteroid.Health,
			Modifiers:  asteroid.DamageModifiers,
		},
		Spec: normalizeRadialDamageSpec(hit.Damage),
	}
}

func radialDamageRequestFromHitAndEnemy(hit radial.Hit, enemy *runtime.Ship) damage.DamageResolutionRequest {
	return damage.DamageResolutionRequest{
		Source: damage.DamageSource{
			EntityID:   hit.SourceID,
			EntityType: damage.EntityTypeProjectile,
			Cause:      damage.DamageCauseArea,
		},
		Target: damage.DamageTarget{
			EntityID:   enemy.ID,
			EntityType: damage.EntityTypeEnemy,
			Health:     enemy.Health,
			Shield:     enemy.Shields,
			Modifiers:  enemy.DamageModifiers,
		},
		Spec: normalizeRadialDamageSpec(hit.Damage),
	}
}

func radialDamageRequestFromHitAndPlayer(hit radial.Hit, player *runtime.Ship) damage.DamageResolutionRequest {
	return damage.DamageResolutionRequest{
		Source: damage.DamageSource{
			EntityID:   hit.SourceID,
			EntityType: damage.EntityTypeProjectile,
			Cause:      damage.DamageCauseArea,
		},
		Target: damage.DamageTarget{
			EntityID:   player.ID,
			EntityType: damage.EntityTypePlayer,
			Health:     player.Health,
			Shield:     player.Shields,
			Modifiers:  player.DamageModifiers,
		},
		Spec: normalizeRadialDamageSpec(hit.Damage),
	}
}

func normalizeRadialDamageSpec(spec damage.DamageSpec) damage.DamageSpec {
	if spec.Cause == "" {
		spec.Cause = damage.DamageCauseArea
	}
	return spec
}
