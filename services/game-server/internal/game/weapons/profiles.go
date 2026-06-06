package weapons

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/effects/radial"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
)

type ImpactEffectKind string

const (
	ImpactEffectNone   ImpactEffectKind = "none"
	ImpactEffectRadial ImpactEffectKind = "radial"
)

type ImpactEffectSpec struct {
	Kind   ImpactEffectKind
	Radial radial.Spec
}

type ProjectileProfile struct {
	Type        string
	Speed       float64
	Lifetime    float64
	SpawnOffset float64
}

type Profile struct {
	ID               ID
	Slot             Slot
	CooldownSeconds   float64
	Projectile       ProjectileProfile
	Damage           damage.DamageSpec
	ImpactEffect     ImpactEffectSpec
}

func Lookup(id ID) (Profile, bool) {
	switch id {
	case BasicCannon:
		return Profile{
			ID:              BasicCannon,
			Slot:            Primary,
			CooldownSeconds: constants.BasicCannonCooldown,
			Projectile: ProjectileProfile{
				Type:        "bullet",
				Speed:       constants.BasicCannonProjectileSpeed,
				Lifetime:    constants.BasicCannonProjectileLifetime,
				SpawnOffset: constants.BasicCannonProjectileSpawnOffset,
			},
			Damage: damage.DamageSpec{
				Amount: constants.BasicCannonDamage,
				Type:   damage.DamageTypeKinetic,
				Cause:  damage.DamageCauseProjectile,
			},
		}, true
	case Torpedo:
		return Profile{
			ID:              Torpedo,
			Slot:            Secondary,
			CooldownSeconds: constants.TorpedoCooldown,
			Projectile: ProjectileProfile{
				Type:        "torpedo",
				Speed:       constants.TorpedoProjectileSpeed,
				Lifetime:    constants.TorpedoProjectileLifetime,
				SpawnOffset: constants.TorpedoProjectileSpawnOffset,
			},
			Damage: damage.DamageSpec{
				Amount: constants.TorpedoImpactDamage,
				Type:   damage.DamageTypeExplosive,
				Cause:  damage.DamageCauseProjectile,
			},
			ImpactEffect: ImpactEffectSpec{
				Kind: ImpactEffectRadial,
				Radial: radial.Spec{
					CoverageMode:       radial.CoverageAnnularWave,
					ExpirationMode:     radial.ExpirationSimultaneous,
					TargetFilter:       radial.TargetFilter{Asteroids: true, Enemies: true},
					ZoneCount:          constants.TorpedoRadialZoneCount,
					ZoneWidth:          constants.TorpedoRadialZoneWidth,
					ZoneSpawnSeconds:   constants.TorpedoRadialZoneSpawnSeconds,
					TickSeconds:        constants.TorpedoRadialTickSeconds,
					TotalSeconds:       constants.TorpedoRadialTotalSeconds,
					ZoneLifetimeSeconds: constants.TorpedoRadialZoneLifetimeSeconds,
					Damage: damage.DamageSpec{
						Amount: constants.TorpedoRadialDamage,
						Type:   damage.DamageTypeExplosive,
						Cause:  damage.DamageCauseArea,
					},
				},
			},
		}, true
	default:
		return Profile{}, false
	}
}
