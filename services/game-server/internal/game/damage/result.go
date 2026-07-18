package damage

type DamageResultKind string

const (
	DamageResultKindDamage                DamageResultKind = "damage"
	DamageResultKindHealing               DamageResultKind = "healing"
	DamageResultKindRepair                DamageResultKind = "repair"
	DamageResultKindBlocked               DamageResultKind = "blocked"
	DamageResultKindIneffective           DamageResultKind = "ineffective"
	DamageResultKindDiscardedLethalTarget DamageResultKind = "discarded_lethal_target"
)

type DamageResult struct {
	TargetEntityID        string
	TargetEntityType      EntityType
	SourceEntityID        string
	SourceEntityType      EntityType
	Kind                  DamageResultKind
	BaseAmount            int
	ModifiedAmount        int
	Type                  DamageType
	Cause                 DamageCause
	AppliedModifiers      []AppliedDamageModifier
	AppliedToHealth       int
	AbsorbedByShield      int
	RestoredToHealth      int
	RestoredToShield      int
	Ignored               bool
	Discarded             bool
	Destroyed             bool
	Fatal                 bool
	RemainingHealth       int
	RemainingShield       int
	CreatedDamageOverTime []ActiveDamageOverTime
	Reason                string
}
