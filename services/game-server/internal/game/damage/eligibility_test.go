package damage

import "testing"

func TestStandardDamagePolicyIdentifier(t *testing.T) {
	policy := NewStandardPolicy()
	if policy.ID != PolicyStandardV1 {
		t.Fatalf("policy ID = %q, want %q", policy.ID, PolicyStandardV1)
	}
	if policy.PvPEnabled {
		t.Fatal("standard policy should default PvP to disabled")
	}
}

func TestEvaluateEligibilityBlocksPlayerFriendlyFireEvenWhenExplicitlyPermitted(t *testing.T) {
	result := EvaluateEligibility(EligibilityRequest{
		Policy:                   Policy{ID: PolicyStandardV1, PvPEnabled: true},
		Source:                   DamageSource{Cause: DamageCauseArea},
		Target:                   DamageTarget{EntityType: EntityTypePlayer},
		Relationship:             RelationshipAlly,
		SourceIsPlayerControlled: true,
		UseExplicitPermissions:   true,
		Permissions:              RelationshipPermissions{Allies: true},
	})
	if result.Eligible || result.Reason != EligibilityBlockPlayerFriendlyFire {
		t.Fatalf("eligibility = %+v", result)
	}
}

func TestEvaluateEligibilityRequiresPvPForOpposingPlayers(t *testing.T) {
	request := EligibilityRequest{
		Policy:                   NewStandardPolicy(),
		Source:                   DamageSource{Cause: DamageCauseProjectile},
		Target:                   DamageTarget{EntityType: EntityTypePlayer},
		Relationship:             RelationshipEnemy,
		SourceIsPlayerControlled: true,
	}
	blocked := EvaluateEligibility(request)
	if blocked.Eligible || blocked.Reason != EligibilityBlockPvPDisabled {
		t.Fatalf("blocked eligibility = %+v", blocked)
	}
	request.Policy.PvPEnabled = true
	allowed := EvaluateEligibility(request)
	if !allowed.Eligible || allowed.Reason != EligibilityBlockNone {
		t.Fatalf("allowed eligibility = %+v", allowed)
	}
}

func TestEvaluateEligibilityInvulnerabilityBypassRequiresAuthorizedDevAdmin(t *testing.T) {
	request := EligibilityRequest{
		Policy:                NewStandardPolicy(),
		Source:                DamageSource{Cause: DamageCauseDebug},
		Target:                DamageTarget{EntityType: EntityTypePlayer},
		Relationship:          RelationshipSelf,
		TargetInvulnerable:    true,
		BypassInvulnerability: true,
	}
	unauthorized := EvaluateEligibility(request)
	if unauthorized.Eligible || unauthorized.Reason != EligibilityBlockUnauthorizedBypass {
		t.Fatalf("unauthorized eligibility = %+v", unauthorized)
	}
	request.AuthorizedDevAdminSource = true
	authorized := EvaluateEligibility(request)
	if !authorized.Eligible {
		t.Fatalf("authorized eligibility = %+v", authorized)
	}
}

func TestDefaultRelationshipPermissions(t *testing.T) {
	projectile := DefaultRelationshipPermissions(DamageSource{Cause: DamageCauseProjectile})
	if projectile.Self || projectile.Allies || !projectile.Enemies || !projectile.Neutrals || !projectile.Destructible {
		t.Fatalf("projectile permissions = %+v", projectile)
	}
	area := DefaultRelationshipPermissions(DamageSource{Cause: DamageCauseArea})
	if !area.Self || area.Allies {
		t.Fatalf("area permissions = %+v", area)
	}
	debug := DefaultRelationshipPermissions(DamageSource{Cause: DamageCauseDebug})
	if !debug.Self || !debug.Allies {
		t.Fatalf("debug permissions = %+v", debug)
	}
}
