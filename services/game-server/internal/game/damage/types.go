package damage

type EntityType string

const (
	EntityTypePlayer     EntityType = "player"
	EntityTypeAsteroid   EntityType = "asteroid"
	EntityTypeProjectile EntityType = "projectile"
)

type DamageType string

type DamageCause string

const (
	DamageCauseCollision  DamageCause = "collision"
	DamageCauseProjectile DamageCause = "projectile"
	DamageCauseDebug      DamageCause = "debug"
	DamageCauseArea       DamageCause = "area"
	DamageCauseDot        DamageCause = "dot"
)

type DamageKind string

const (
	DamageKindKinetic   DamageKind = "kinetic"
	DamageKindExplosive DamageKind = "explosive"
	DamageKindEnergy    DamageKind = "energy"
	DamageKindFire      DamageKind = "fire"
	DamageKindPoison    DamageKind = "poison"
	DamageKindTrueDamage DamageKind = "true_damage"
)
