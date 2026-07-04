package quantize

import "testing"

func TestLookupPolicyEventFieldPolicies(t *testing.T) {
	tests := []struct {
		path       string
		wantPolicy PolicyName
	}{
		{path: "event.bullet_blast.x", wantPolicy: PolicyPosition},
		{path: "event.bullet_blast.y", wantPolicy: PolicyPosition},
		{path: "event.ship_death.x", wantPolicy: PolicyPosition},
		{path: "event.ship_death.y", wantPolicy: PolicyPosition},
		{path: "event.ship_death.respawn_delay", wantPolicy: PolicySeconds},
		{path: "event.damage_applied.x", wantPolicy: PolicyPosition},
		{path: "event.damage_applied.y", wantPolicy: PolicyPosition},
		{path: "event.radial_effect_started.x", wantPolicy: PolicyPosition},
		{path: "event.radial_effect_started.y", wantPolicy: PolicyPosition},
		{path: "event.pickup_collected.x", wantPolicy: PolicyPosition},
		{path: "event.pickup_collected.y", wantPolicy: PolicyPosition},
		{path: "event.pickup_expired.x", wantPolicy: PolicyPosition},
		{path: "event.pickup_expired.y", wantPolicy: PolicyPosition},
		{path: "event.pickup_dropped.x", wantPolicy: PolicyPosition},
		{path: "event.pickup_dropped.y", wantPolicy: PolicyPosition},
		{path: "event.damage_over_time_tick.x", wantPolicy: PolicyPosition},
		{path: "event.damage_over_time_tick.y", wantPolicy: PolicyPosition},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			policy, ok := LookupPolicy(tc.path)
			if !ok {
				t.Fatalf("LookupPolicy(%q) did not find an explicit policy", tc.path)
			}
			if policy.Name != tc.wantPolicy {
				t.Fatalf("LookupPolicy(%q) = %q, want %q", tc.path, policy.Name, tc.wantPolicy)
			}
			if policy.Name == PolicyFloatGeneric {
				t.Fatalf("LookupPolicy(%q) fell back to float_generic", tc.path)
			}
		})
	}
}
