package damage

type DamageSource struct {
	EntityID                 string
	EntityType               EntityType
	Cause                    DamageCause
	ResponsiblePlayerID      string
	ResponsibleTeamID        string
	OriginalInstigatorID     string
	Permissions              RelationshipPermissions
	UseExplicitPermissions   bool
	BypassInvulnerability    bool
	AuthorizedDevAdminSource bool
}

type DamageTarget struct {
	EntityID     string
	EntityType   EntityType
	Health       int
	MaxHealth    int
	Shield       int
	MaxShield    int
	Invulnerable bool
	Modifiers    []DamageModifier
}

type RestorationDestination string

const (
	RestorationDestinationHealth RestorationDestination = "health"
	RestorationDestinationShield RestorationDestination = "shield"
	RestorationDestinationBoth   RestorationDestination = "both"
)

type ShieldOverflowPolicy string

const (
	ShieldOverflowPassThrough ShieldOverflowPolicy = "pass_through"
	ShieldOverflowDiscard     ShieldOverflowPolicy = "discard"

	DamageOverflowPassThrough = ShieldOverflowPassThrough
	DamageOverflowDiscard     = ShieldOverflowDiscard
)

type DamageSpec struct {
	Amount                 int
	Type                   DamageType
	Cause                  DamageCause
	BypassShield           bool
	OverflowPolicy         ShieldOverflowPolicy
	RestorationDestination RestorationDestination
	DoT                    DamageOverTimeSpec
}

type DamageResolutionRequest struct {
	Source    DamageSource
	Target    DamageTarget
	Spec      DamageSpec
	Modifiers []DamageModifier
}
