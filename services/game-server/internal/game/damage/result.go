package damage

type DamageResult struct {
	TargetEntityID   string
	TargetEntityType EntityType
	SourceEntityID   string
	SourceEntityType EntityType
	BaseAmount       int
	ModifiedAmount   int
	Kind             DamageKind
	Cause            DamageCause
	AppliedModifiers []AppliedDamageModifier
	AppliedToHealth  int
	AbsorbedByShield int
	Ignored          bool
	Destroyed        bool
	Fatal            bool
	RemainingHealth  int
	RemainingShield  int
	CreatedDamageOverTime []ActiveDamageOverTime
	Reason           string
}
