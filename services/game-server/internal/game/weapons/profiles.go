package weapons

import (
	"github.com/Lokee86/space-rocks/server/internal/constants"
	"github.com/Lokee86/space-rocks/server/internal/game/damage"
)

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
	default:
		return Profile{}, false
	}
}
