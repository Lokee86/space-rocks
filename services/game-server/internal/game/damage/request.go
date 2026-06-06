package damage

type DamageSource struct {
	EntityID   string
	EntityType EntityType
	Cause      DamageCause
}

type DamageTarget struct {
	EntityID   string
	EntityType EntityType
	Health     int
	Shield     int
	Modifiers  []DamageModifier
}

type DamageSpec struct {
	Amount       int
	Kind         DamageKind
	Cause        DamageCause
	BypassShield bool
	DoT          DamageOverTimeSpec
}

type DamageResolutionRequest struct {
	Source    DamageSource
	Target    DamageTarget
	Spec      DamageSpec
	Modifiers []DamageModifier
}
