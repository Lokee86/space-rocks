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
			ID:             BasicCannon,
			Slot:           Primary,
			CooldownSeconds: constants.BulletCooldown,
			Projectile: ProjectileProfile{
				Type:        "bullet",
				Speed:       constants.BulletSpeed,
				Lifetime:    constants.BulletLifetime,
				SpawnOffset: constants.BulletSpawnOffset,
			},
			Damage: damage.DamageSpec{
				Amount: constants.BulletDamage,
				Type:   damage.DamageTypeKinetic,
				Cause:  damage.DamageCauseProjectile,
			},
		}, true
	case Torpedo:
		return Profile{
			ID:             Torpedo,
			Slot:           Secondary,
			CooldownSeconds: constants.BulletCooldown,
			Projectile: ProjectileProfile{
				Type:        "torpedo",
				Speed:       constants.BulletSpeed,
				Lifetime:    constants.BulletLifetime,
				SpawnOffset: constants.BulletSpawnOffset,
			},
			Damage: damage.DamageSpec{
				Amount: 2,
				Type:   damage.DamageTypeExplosive,
				Cause:  damage.DamageCauseProjectile,
			},
			ImpactEffect: ImpactEffectSpec{
				Kind: ImpactEffectRadial,
				Radial: radial.Spec{
					CoverageMode:       radial.CoverageAnnularWave,
					ExpirationMode:     radial.ExpirationSimultaneous,
					TargetFilter:       radial.TargetFilter{Asteroids: true, Enemies: true},
					ZoneCount:          4,
					ZoneWidth:          10,
					ZoneSpawnSeconds:   0.1,
					TickSeconds:        0.1,
					TotalSeconds:       0.4,
					ZoneLifetimeSeconds: 0.4,
					Damage: damage.DamageSpec{
						Amount: 2,
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
