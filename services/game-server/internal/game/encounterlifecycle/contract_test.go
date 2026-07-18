package encounterlifecycle

import "testing"

func validOrigin() OriginMetadata {
	return OriginMetadata{
		ProfileID:              "profile-1",
		SpawnType:              "ambush",
		LifecyclePolicyID:      "standard",
		Priority:               2,
		WeightedPopulationCost: 3,
		PhaseID:                "phase-1",
	}
}

func TestValidateOriginAcceptsOptionalPhase(t *testing.T) {
	metadata := validOrigin()
	if err := ValidateOrigin(metadata); err != nil {
		t.Fatalf("expected valid origin, got %v", err)
	}
	if !metadata.HasPhase() {
		t.Fatal("expected phase association")
	}

	metadata.PhaseID = ""
	if err := ValidateOrigin(metadata); err != nil {
		t.Fatalf("expected origin without phase to be valid, got %v", err)
	}
	if metadata.HasPhase() {
		t.Fatal("expected empty phase ID to be absent")
	}
}

func TestDecideAcceptsSupportedTriggerFamilies(t *testing.T) {
	triggers := []Trigger{
		TriggerLifetimeExpiry,
		TriggerOutsideAllRelevantPlayers,
		TriggerAllowedRegionExit,
		TriggerPopulationPressure,
		TriggerProfilePhaseCleanup,
		TriggerScriptedCleanup,
	}
	for _, trigger := range triggers {
		decision, err := Decide(DecisionRequest{
			Origin:               validOrigin(),
			Trigger:              trigger,
			Capabilities:         EntityCapabilities{SupportsSoftRetire: true},
			RequestedDisposition: DispositionSoftRetire,
		})
		if err != nil {
			t.Fatalf("trigger %q rejected: %v", trigger, err)
		}
		if decision.Disposition != DispositionSoftRetire {
			t.Fatalf("trigger %q returned %q", trigger, decision.Disposition)
		}
	}

	decision, err := Decide(DecisionRequest{
		Origin:               validOrigin(),
		Trigger:              TriggerTransitionReset,
		Capabilities:         EntityCapabilities{SupportsHardRemove: true},
		RequestedDisposition: DispositionHardRemove,
	})
	if err != nil {
		t.Fatalf("transition/reset rejected: %v", err)
	}
	if decision.Disposition != DispositionHardRemove {
		t.Fatalf("expected hard removal, got %q", decision.Disposition)
	}
}

func TestValidateRequestRejectsInvalidContracts(t *testing.T) {
	tests := []struct {
		name    string
		request DecisionRequest
	}{
		{
			name: "missing profile",
			request: DecisionRequest{
				Origin:               OriginMetadata{SpawnType: "ambush", LifecyclePolicyID: "standard", WeightedPopulationCost: 1},
				Trigger:              TriggerLifetimeExpiry,
				Capabilities:         EntityCapabilities{SupportsSoftRetire: true},
				RequestedDisposition: DispositionSoftRetire,
			},
		},
		{
			name: "non-positive population cost",
			request: DecisionRequest{
				Origin:               OriginMetadata{ProfileID: "profile-1", SpawnType: "ambush", LifecyclePolicyID: "standard"},
				Trigger:              TriggerLifetimeExpiry,
				Capabilities:         EntityCapabilities{SupportsSoftRetire: true},
				RequestedDisposition: DispositionSoftRetire,
			},
		},
		{
			name: "unknown trigger",
			request: DecisionRequest{
				Origin:               validOrigin(),
				Trigger:              Trigger("unknown"),
				Capabilities:         EntityCapabilities{SupportsSoftRetire: true},
				RequestedDisposition: DispositionSoftRetire,
			},
		},
		{
			name: "unsupported soft retirement",
			request: DecisionRequest{
				Origin:               validOrigin(),
				Trigger:              TriggerLifetimeExpiry,
				Capabilities:         EntityCapabilities{SupportsHardRemove: true},
				RequestedDisposition: DispositionSoftRetire,
			},
		},
		{
			name: "required destruction with soft retirement",
			request: DecisionRequest{
				Origin:               validOrigin(),
				Trigger:              TriggerLifetimeExpiry,
				Capabilities:         EntityCapabilities{SupportsSoftRetire: true, SupportsHardRemove: true, RequiresDestruction: true},
				RequestedDisposition: DispositionSoftRetire,
			},
		},
		{
			name: "conflicting capabilities",
			request: DecisionRequest{
				Origin:               validOrigin(),
				Trigger:              TriggerLifetimeExpiry,
				Capabilities:         EntityCapabilities{SupportsSoftRetire: true, SupportsHardRemove: true, RequiresExplicitCleanup: true, RequiresDestruction: true},
				RequestedDisposition: DispositionSoftRetire,
			},
		},
		{
			name: "transition reset soft retirement",
			request: DecisionRequest{
				Origin:               validOrigin(),
				Trigger:              TriggerTransitionReset,
				Capabilities:         EntityCapabilities{SupportsSoftRetire: true},
				RequestedDisposition: DispositionSoftRetire,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := ValidateRequest(test.request); err == nil {
				t.Fatal("expected invalid request to be rejected")
			}
		})
	}
}
