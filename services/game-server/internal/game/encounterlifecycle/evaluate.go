package encounterlifecycle

// RelevantPlayerDistance contains a wrap-aware distance already measured by
// the runtime for one active relevant player.
type RelevantPlayerDistance struct {
	Distance       float64
	ViewableRadius float64
}

// EvaluationFacts are authoritative runtime facts without a Game dependency.
type EvaluationFacts struct {
	ElapsedLifetimeSeconds float64
	RelevantPlayers        []RelevantPlayerDistance
	InsideAllowedRegion    bool

	PopulationCleanupRequested   bool
	ProfilePhaseCleanupRequested bool
	ScriptedCleanupRequested     bool
	TransitionResetRequested     bool
	SimulationPaused             bool
}

// EvaluationRequest is the pure lifecycle policy evaluation boundary.
type EvaluationRequest struct {
	Origin       OriginMetadata
	Capabilities EntityCapabilities
	Policy       Policy
	Facts        EvaluationFacts
}

// EvaluationResult identifies the selected trigger and validated disposition.
type EvaluationResult struct {
	Trigger     Trigger
	Disposition Disposition
}

// Evaluate selects at most one lifecycle trigger using the explicit precedence
// order below, then reuses Decide for capability and disposition validation.
func Evaluate(request EvaluationRequest) (EvaluationResult, bool, error) {
	if request.Facts.SimulationPaused {
		return EvaluationResult{}, false, nil
	}
	if err := request.Policy.Validate(); err != nil {
		return EvaluationResult{}, false, err
	}

	trigger, disposition, triggered := firstTriggered(request)
	if !triggered {
		return EvaluationResult{}, false, nil
	}

	decision, err := Decide(DecisionRequest{
		Origin:               request.Origin,
		Trigger:              trigger,
		Capabilities:         request.Capabilities,
		RequestedDisposition: disposition,
	})
	if err != nil {
		return EvaluationResult{}, false, err
	}
	return EvaluationResult{Trigger: trigger, Disposition: decision.Disposition}, true, nil
}

func firstTriggered(request EvaluationRequest) (Trigger, Disposition, bool) {
	// Precedence is explicit and deterministic: transition/reset, lifetime,
	// outside all players, allowed region, population, profile/phase, scripted.
	if request.Policy.TransitionReset.Enabled && request.Facts.TransitionResetRequested {
		return TriggerTransitionReset, request.Policy.TransitionReset.Disposition, true
	}
	if request.Policy.LifetimeExpiry.Enabled && request.Facts.ElapsedLifetimeSeconds >= request.Policy.LifetimeSeconds {
		return TriggerLifetimeExpiry, request.Policy.LifetimeExpiry.Disposition, true
	}
	if request.Policy.OutsideAllRelevantPlayers.Enabled && outsideAllRelevantPlayers(request.Facts.RelevantPlayers, request.Policy.ExtraPlayerDistance) {
		return TriggerOutsideAllRelevantPlayers, request.Policy.OutsideAllRelevantPlayers.Disposition, true
	}
	if request.Policy.AllowedRegionExit.Enabled && !request.Facts.InsideAllowedRegion {
		return TriggerAllowedRegionExit, request.Policy.AllowedRegionExit.Disposition, true
	}
	if request.Policy.PopulationPressure.Enabled && request.Facts.PopulationCleanupRequested {
		return TriggerPopulationPressure, request.Policy.PopulationPressure.Disposition, true
	}
	if request.Policy.ProfilePhaseCleanup.Enabled && request.Facts.ProfilePhaseCleanupRequested {
		return TriggerProfilePhaseCleanup, request.Policy.ProfilePhaseCleanup.Disposition, true
	}
	if request.Policy.ScriptedCleanup.Enabled && request.Facts.ScriptedCleanupRequested {
		return TriggerScriptedCleanup, request.Policy.ScriptedCleanup.Disposition, true
	}
	return "", "", false
}

func outsideAllRelevantPlayers(players []RelevantPlayerDistance, extraDistance float64) bool {
	if len(players) == 0 {
		return false
	}
	for _, player := range players {
		if player.Distance <= player.ViewableRadius+extraDistance {
			return false
		}
	}
	return true
}
