package damage

type EntityType string

const (
	EntityTypePlayer     EntityType = "player"
	EntityTypeAsteroid   EntityType = "asteroid"
	EntityTypeProjectile EntityType = "projectile"
)

type DamageCause string

const (
	DamageCauseCollision  DamageCause = "collision"
	DamageCauseProjectile DamageCause = "projectile"
	DamageCauseDebug      DamageCause = "debug"
	DamageCauseArea       DamageCause = "area"
	DamageCauseDot        DamageCause = "dot"
)

type DamageType string

const (
	DamageTypeKinetic     DamageType = "kinetic"
	DamageTypeExplosive   DamageType = "explosive"
	DamageTypeEnergy      DamageType = "energy"
	DamageTypeThermal     DamageType = "thermal"
	DamageTypeRadioactive DamageType = "radioactive"
	DamageTypeTrueDamage  DamageType = "true_damage"
)
