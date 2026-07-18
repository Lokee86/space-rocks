package damage

type PolicyID string

const PolicyStandardV1 PolicyID = "standard_damage_v1"

type Policy struct {
	ID         PolicyID
	PvPEnabled bool
}

func NewStandardPolicy() Policy {
	return Policy{ID: PolicyStandardV1}
}

type Relationship string

const (
	RelationshipSelf         Relationship = "self"
	RelationshipAlly         Relationship = "ally"
	RelationshipEnemy        Relationship = "enemy"
	RelationshipNeutral      Relationship = "neutral"
	RelationshipDestructible Relationship = "destructible"
)

type RelationshipPermissions struct {
	Self         bool
	Allies       bool
	Enemies      bool
	Neutrals     bool
	Destructible bool
}

func DefaultRelationshipPermissions(source DamageSource) RelationshipPermissions {
	permissions := RelationshipPermissions{
		Enemies:      true,
		Neutrals:     true,
		Destructible: true,
	}
	if source.Cause == DamageCauseArea || source.Cause == DamageCauseDebug {
		permissions.Self = true
	}
	if source.Cause == DamageCauseDebug {
		permissions.Allies = true
	}
	return permissions
}
