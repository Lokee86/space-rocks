package encounterlifecycle

import "testing"

func TestEvaluateSelectsEveryTriggerFamily(t *testing.T) {
	tests := []struct {
		name         string
		policy       Policy
		facts        EvaluationFacts
		capabilities EntityCapabilities
		wantTrigger  Trigger
		want         Disposition
	}{
		{
			name:         "lifetime expiry",
			policy:       Policy{LifetimeExpiry: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire}, LifetimeSeconds: 5},
			facts:        EvaluationFacts{ElapsedLifetimeSeconds: 5},
			capabilities: EntityCapabilities{SupportsSoftRetire: true},
			wantTrigger:  TriggerLifetimeExpiry,
			want:         DispositionSoftRetire,
		},
		{
			name: "outside all relevant players",
			policy: Policy{
				OutsideAllRelevantPlayers: TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove},
				ExtraPlayerDistance:       5,
			},
			facts:        EvaluationFacts{RelevantPlayers: []RelevantPlayerDistance{{Distance: 20, ViewableRadius: 10}, {Distance: 30, ViewableRadius: 10}}},
			capabilities: EntityCapabilities{SupportsHardRemove: true},
			wantTrigger:  TriggerOutsideAllRelevantPlayers,
			want:         DispositionHardRemove,
		},
		{
			name:         "allowed region exit",
			policy:       Policy{AllowedRegionExit: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire}},
			facts:        EvaluationFacts{InsideAllowedRegion: false},
			capabilities: EntityCapabilities{SupportsSoftRetire: true},
			wantTrigger:  TriggerAllowedRegionExit,
			want:         DispositionSoftRetire,
		},
		{
			name:         "population pressure",
			policy:       Policy{PopulationPressure: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire}},
			facts:        EvaluationFacts{PopulationCleanupRequested: true},
			capabilities: EntityCapabilities{SupportsSoftRetire: true},
			wantTrigger:  TriggerPopulationPressure,
			want:         DispositionSoftRetire,
		},
		{
			name:         "profile phase cleanup",
			policy:       Policy{ProfilePhaseCleanup: TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove}},
			facts:        EvaluationFacts{ProfilePhaseCleanupRequested: true},
			capabilities: EntityCapabilities{SupportsHardRemove: true},
			wantTrigger:  TriggerProfilePhaseCleanup,
			want:         DispositionHardRemove,
		},
		{
			name:         "scripted cleanup",
			policy:       Policy{ScriptedCleanup: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire}},
			facts:        EvaluationFacts{ScriptedCleanupRequested: true},
			capabilities: EntityCapabilities{SupportsSoftRetire: true},
			wantTrigger:  TriggerScriptedCleanup,
			want:         DispositionSoftRetire,
		},
		{
			name:         "transition reset",
			policy:       Policy{TransitionReset: TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove}},
			facts:        EvaluationFacts{TransitionResetRequested: true},
			capabilities: EntityCapabilities{SupportsHardRemove: true},
			wantTrigger:  TriggerTransitionReset,
			want:         DispositionHardRemove,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, decided, err := Evaluate(EvaluationRequest{
				Origin:       validOrigin(),
				Capabilities: test.capabilities,
				Policy:       test.policy,
				Facts:        test.facts,
			})
			if err != nil {
				t.Fatalf("evaluation failed: %v", err)
			}
			if !decided {
				t.Fatal("expected a lifecycle decision")
			}
			if result.Trigger != test.wantTrigger || result.Disposition != test.want {
				t.Fatalf("got trigger %q and disposition %q", result.Trigger, result.Disposition)
			}
		})
	}
}

func TestEvaluateOutsidePlayerRequiresEveryPlayerToBeOutside(t *testing.T) {
	request := EvaluationRequest{
		Origin:       validOrigin(),
		Capabilities: EntityCapabilities{SupportsSoftRetire: true},
		Policy: Policy{
			OutsideAllRelevantPlayers: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
			ExtraPlayerDistance:       5,
		},
		Facts: EvaluationFacts{RelevantPlayers: []RelevantPlayerDistance{
			{Distance: 20, ViewableRadius: 10},
			{Distance: 14, ViewableRadius: 10},
		}},
	}
	if _, decided, err := Evaluate(request); err != nil || decided {
		t.Fatalf("expected nearby player to block retirement, decided=%v err=%v", decided, err)
	}

	request.Facts.RelevantPlayers[1].Distance = 16
	if _, decided, err := Evaluate(request); err != nil || !decided {
		t.Fatalf("expected all players outside to trigger retirement, decided=%v err=%v", decided, err)
	}

	request.Facts.RelevantPlayers = nil
	if _, decided, err := Evaluate(request); err != nil || decided {
		t.Fatalf("expected zero relevant players not to trigger distance retirement, decided=%v err=%v", decided, err)
	}
}

func TestEvaluatePauseProducesNoDecision(t *testing.T) {
	result, decided, err := Evaluate(EvaluationRequest{
		Origin:       validOrigin(),
		Capabilities: EntityCapabilities{SupportsHardRemove: true},
		Policy:       Policy{TransitionReset: TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove}},
		Facts:        EvaluationFacts{TransitionResetRequested: true, SimulationPaused: true},
	})
	if err != nil || decided || result != (EvaluationResult{}) {
		t.Fatalf("expected pause to produce no decision, result=%+v decided=%v err=%v", result, decided, err)
	}
}

func TestEvaluateUsesDeterministicTriggerPrecedence(t *testing.T) {
	policy := Policy{
		LifetimeExpiry:            TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		OutsideAllRelevantPlayers: TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		AllowedRegionExit:         TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		PopulationPressure:        TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		ProfilePhaseCleanup:       TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		ScriptedCleanup:           TriggerPolicy{Enabled: true, Disposition: DispositionSoftRetire},
		TransitionReset:           TriggerPolicy{Enabled: true, Disposition: DispositionHardRemove},
		LifetimeSeconds:           1,
	}
	request := EvaluationRequest{
		Origin:       validOrigin(),
		Capabilities: EntityCapabilities{SupportsSoftRetire: true, SupportsHardRemove: true},
		Policy:       policy,
		Facts: EvaluationFacts{
			ElapsedLifetimeSeconds:       2,
			RelevantPlayers:              []RelevantPlayerDistance{{Distance: 100, ViewableRadius: 1}},
			PopulationCleanupRequested:   true,
			ProfilePhaseCleanupRequested: true,
			ScriptedCleanupRequested:     true,
			TransitionResetRequested:     true,
		},
	}

	result, decided, err := Evaluate(request)
	if err != nil || !decided || result.Trigger != TriggerTransitionReset {
		t.Fatalf("expected transition/reset first, result=%+v decided=%v err=%v", result, decided, err)
	}

	request.Facts.TransitionResetRequested = false
	result, decided, err = Evaluate(request)
	if err != nil || !decided || result.Trigger != TriggerLifetimeExpiry {
		t.Fatalf("expected lifetime second, result=%+v decided=%v err=%v", result, decided, err)
	}
}
