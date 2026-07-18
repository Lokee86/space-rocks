package damage

type EligibilityBlockReason string

const (
	EligibilityBlockNone               EligibilityBlockReason = ""
	EligibilityBlockInvulnerable       EligibilityBlockReason = "invulnerable"
	EligibilityBlockRelationship       EligibilityBlockReason = "relationship_not_permitted"
	EligibilityBlockPlayerFriendlyFire EligibilityBlockReason = "player_friendly_fire_prohibited"
	EligibilityBlockPvPDisabled        EligibilityBlockReason = "pvp_disabled"
	EligibilityBlockUnauthorizedBypass EligibilityBlockReason = "unauthorized_invulnerability_bypass"
)

type EligibilityRequest struct {
	Policy                   Policy
	Source                   DamageSource
	Target                   DamageTarget
	Relationship             Relationship
	Permissions              RelationshipPermissions
	UseExplicitPermissions   bool
	SourceIsPlayerControlled bool
	TargetInvulnerable       bool
	BypassInvulnerability    bool
	AuthorizedDevAdminSource bool
}

type EligibilityResult struct {
	Eligible bool
	Reason   EligibilityBlockReason
}

func EvaluateEligibility(request EligibilityRequest) EligibilityResult {
	if request.TargetInvulnerable {
		if !request.BypassInvulnerability {
			return EligibilityResult{Reason: EligibilityBlockInvulnerable}
		}
		if !request.AuthorizedDevAdminSource {
			return EligibilityResult{Reason: EligibilityBlockUnauthorizedBypass}
		}
	}

	if request.AuthorizedDevAdminSource {
		return EligibilityResult{Eligible: true}
	}

	if request.SourceIsPlayerControlled && request.Target.EntityType == EntityTypePlayer {
		if request.Relationship == RelationshipAlly {
			return EligibilityResult{Reason: EligibilityBlockPlayerFriendlyFire}
		}
		if request.Relationship == RelationshipEnemy && !request.Policy.PvPEnabled {
			return EligibilityResult{Reason: EligibilityBlockPvPDisabled}
		}
	}

	permissions := request.Permissions
	if !request.UseExplicitPermissions {
		permissions = DefaultRelationshipPermissions(request.Source)
	}
	if !relationshipPermitted(permissions, request.Relationship) {
		return EligibilityResult{Reason: EligibilityBlockRelationship}
	}
	return EligibilityResult{Eligible: true}
}

func relationshipPermitted(permissions RelationshipPermissions, relationship Relationship) bool {
	switch relationship {
	case RelationshipSelf:
		return permissions.Self
	case RelationshipAlly:
		return permissions.Allies
	case RelationshipEnemy:
		return permissions.Enemies
	case RelationshipNeutral:
		return permissions.Neutrals
	case RelationshipDestructible:
		return permissions.Destructible
	default:
		return false
	}
}
