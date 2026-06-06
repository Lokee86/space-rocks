package weapons

import (
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
	"github.com/Lokee86/space-rocks/server/internal/game/physics"
)

type FireRequest struct {
	Equipped Equipped
	State    SlotState
	Position physics.Vector2
	Forward  physics.Vector2
	Rotation float64
}

type ProjectileSpawn struct {
	WeaponID       ID
	ProjectileType string
	Position       physics.Vector2
	Rotation       float64
	Velocity       physics.Vector2
	Lifetime       float64
	Damage         damage.DamageSpec
	ImpactEffect   ImpactEffectSpec
}

type FireResult struct {
	Fired      bool
	NewState   SlotState
	Projectile ProjectileSpawn
}

func Fire(req FireRequest) FireResult {
	if req.Equipped.ID == "" {
		return FireResult{}
	}

	profile, ok := Lookup(req.Equipped.ID)
	if !ok {
		return FireResult{}
	}

	if req.State.CooldownRemaining > 0 {
		return FireResult{}
	}

	if req.Equipped.AmmoPolicy == LimitedAmmo && req.State.AmmoRemaining <= 0 {
		return FireResult{}
	}

	forward := req.Forward.Normalized()
	spawn := ProjectileSpawn{
		WeaponID:       req.Equipped.ID,
		ProjectileType: profile.Projectile.Type,
		Position:       req.Position.Add(forward.Multiply(profile.Projectile.SpawnOffset)),
		Rotation:       req.Rotation,
		Velocity:       forward.Multiply(profile.Projectile.Speed),
		Lifetime:       profile.Projectile.Lifetime,
		Damage:         profile.Damage,
		ImpactEffect:   profile.ImpactEffect,
	}

	result := FireResult{
		Fired:      true,
		NewState:   req.State,
		Projectile: spawn,
	}
	result.NewState.CooldownRemaining = profile.CooldownSeconds
	if req.Equipped.AmmoPolicy == LimitedAmmo {
		result.NewState.AmmoRemaining--
	}

	return result
}
