package encounterlifecycle

import "testing"

func TestPolicyValidateAcceptsConfiguredTriggerDispositions(t *testing.T) {
	policy := Policy{
		LifetimeExpiry:            TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		OutsideAllRelevantPlayers: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		AllowedRegionExit:         TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove},
		PopulationPressure:        TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		ProfilePhaseCleanup:       TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		ScriptedCleanup:           TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove},
		TransitionReset:           TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove},
		LifetimeSeconds:           10,
		ExtraPlayerDistance:       5,
	}
	if err := policy.Validate(); err != nil {
		t.Fatalf("expected valid policy, got %v", err)
	}
}

func TestPolicyValidateRejectsImpossibleConfiguration(t *testing.T) {
	tests := []Policy{
		{LifetimeExpiry: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire}},
		{OutsideAllRelevantPlayers: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire}, ExtraPlayerDistance: -1},
		{LifetimeExpiry: TriggerPolicy{Enabled: true, Disposition: Disposition("unknown")}, LifetimeSeconds: 1},
		{TransitionReset: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire}},
	}
	for index, policy := range tests {
		if err := policy.Validate(); err == nil {
			t.Fatalf("policy %d was accepted", index)
		}
	}
}
