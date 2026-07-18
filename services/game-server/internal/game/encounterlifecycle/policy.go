package encounterlifecycle

import "fmt"

// TriggerPolicy enables one trigger family and selects its cleanup disposition.
type TriggerPolicy struct {
	Enabled     bool
	Disposition Disposition
}

// Policy contains lifecycle trigger configuration and the thresholds needed by
// the pure evaluation boundary. Distances are supplied as authoritative facts.
type Policy struct {
	LifetimeExpiry            TriggerPolicy
	OutsideAllRelevantPlayers TriggerPolicy
	AllowedRegionExit         TriggerPolicy
	PopulationPressure        TriggerPolicy
	ProfilePhaseCleanup       TriggerPolicy
	ScriptedCleanup           TriggerPolicy
	TransitionReset           TriggerPolicy

	LifetimeSeconds     float64
	ExtraPlayerDistance float64
}

func (policy Policy) Validate() error {
	if policy.LifetimeExpiry.Enabled && policy.LifetimeSeconds <= 0 {
		return fmt.Errorf("enabled lifetime expiry requires a positive lifetime")
	}
	if policy.OutsideAllRelevantPlayers.Enabled && policy.ExtraPlayerDistance < 0 {
		return fmt.Errorf("outside-player extra distance cannot be negative")
	}

	triggerPolicies := []struct {
		trigger Trigger
		policy  TriggerPolicy
	}{
		{TriggerLifetimeExpiry, policy.LifetimeExpiry},
		{TriggerOutsideAllRelevantPlayers, policy.OutsideAllRelevantPlayers},
		{TriggerAllowedRegionExit, policy.AllowedRegionExit},
		{TriggerPopulationPressure, policy.PopulationPressure},
		{TriggerProfilePhaseCleanup, policy.ProfilePhaseCleanup},
		{TriggerScriptedCleanup, policy.ScriptedCleanup},
		{TriggerTransitionReset, policy.TransitionReset},
	}
	for _, configured := range triggerPolicies {
		if !configured.policy.Enabled {
			continue
		}
		if !isSupportedDisposition(configured.policy.Disposition) {
			return fmt.Errorf("%s has unsupported disposition %q", configured.trigger, configured.policy.Disposition)
		}
		if configured.trigger == TriggerTransitionReset && configured.policy.Disposition != DispositionHardRemove {
			return fmt.Errorf("transition/reset policy requires hard removal")
		}
	}
	return nil
}
