package game

import (
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/damage"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/services/game-server/internal/game/runtime"
)

func radialDamageSource(hit radial.Hit) damage.DamageSource {
	return damage.DamageSource{
		EntityID:             hit.SourceID,
		EntityType:           damage.EntityTypeProjectile,
		Cause:                damage.DamageCauseArea,
		ResponsiblePlayerID:  hit.SourcePlayerID,
		OriginalInstigatorID: hit.SourcePlayerID,
	}
}

func radialDamageRequestFromHitAndAsteroid(hit radial.Hit, asteroid *runtime.Asteroid) damage.DamageResolutionRequest {
	return damage.DamageResolutionRequest{
		Source: radialDamageSource(hit),
		Target: damage.DamageTarget{
			EntityID:   asteroid.ID,
			EntityType: damage.EntityTypeAsteroid,
			Health:     asteroid.Health,
			MaxHealth:  asteroid.Health,
			Modifiers:  asteroid.DamageModifiers,
		},
		Spec: normalizeRadialDamageSpec(hit.Damage),
	}
}

func radialDamageRequestFromHitAndEnemy(hit radial.Hit, enemy *runtime.Ship) damage.DamageResolutionRequest {
	return damage.DamageResolutionRequest{
		Source: radialDamageSource(hit),
		Target: damage.DamageTarget{
			EntityID:     enemy.ID,
			EntityType:   damage.EntityTypeEnemy,
			Health:       enemy.Health,
			MaxHealth:    enemy.Stats.MaxHealth,
			Shield:       enemy.Shields,
			MaxShield:    enemy.Stats.MaxShields,
			Invulnerable: enemy.IsInvulnerable() || !enemy.DamageOptions.CanTakeDamage(),
			Modifiers:    enemy.DamageModifiers,
		},
		Spec: normalizeRadialDamageSpec(hit.Damage),
	}
}

func radialDamageRequestFromHitAndPlayer(hit radial.Hit, player *runtime.Ship) damage.DamageResolutionRequest {
	return damage.DamageResolutionRequest{
		Source: radialDamageSource(hit),
		Target: damage.DamageTarget{
			EntityID:     player.ID,
			EntityType:   damage.EntityTypePlayer,
			Health:       player.Health,
			MaxHealth:    player.Stats.MaxHealth,
			Shield:       player.Shields,
			MaxShield:    player.Stats.MaxShields,
			Invulnerable: player.IsInvulnerable() || !player.DamageOptions.CanTakeDamage(),
			Modifiers:    player.DamageModifiers,
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
