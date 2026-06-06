package radial

type TargetKind string

const (
	TargetAsteroid   TargetKind = "asteroid"
	TargetEnemy      TargetKind = "enemy"
	TargetPlayer     TargetKind = "player"
	TargetProjectile TargetKind = "projectile"
	TargetPickup     TargetKind = "pickup"
)

type TargetFilter struct {
	Asteroids   bool
	Enemies     bool
	Players     bool
	Projectiles bool
	Pickups     bool
}

func (filter TargetFilter) Allows(kind TargetKind) bool {
	switch kind {
	case TargetAsteroid:
		return filter.Asteroids
	case TargetEnemy:
		return filter.Enemies
	case TargetPlayer:
		return filter.Players
	case TargetProjectile:
		return filter.Projectiles
	case TargetPickup:
		return filter.Pickups
	default:
		return false
	}
}
