package encounterlifecycle

import "fmt"

// DecisionRequest is the pure input boundary for a lifecycle disposition.
// RequestedDisposition is selected by the active lifecycle policy; this
// package validates that the entity can honor it for the supplied trigger.
type DecisionRequest struct {
	Origin               OriginMetadata
	Trigger              Trigger
	Capabilities         EntityCapabilities
	RequestedDisposition Disposition
}

// Decision is the cleanup handoff selected for a valid request.
type Decision struct {
	Disposition Disposition
}

// Decide validates a lifecycle request and returns its cleanup disposition.
func Decide(request DecisionRequest) (Decision, error) {
	if err := ValidateRequest(request); err != nil {
		return Decision{}, err
	}
	return Decision{Disposition: request.RequestedDisposition}, nil
}

func ValidateRequest(request DecisionRequest) error {
	if err := ValidateOrigin(request.Origin); err != nil {
		return err
	}
	if !isSupportedTrigger(request.Trigger) {
		return fmt.Errorf("unsupported encounter lifecycle trigger %q", request.Trigger)
	}
	if err := ValidateCapabilities(request.Capabilities); err != nil {
		return err
	}
	if !isSupportedDisposition(request.RequestedDisposition) {
		return fmt.Errorf("unsupported encounter lifecycle disposition %q", request.RequestedDisposition)
	}

	if request.Trigger == TriggerTransitionReset && request.RequestedDisposition != DispositionHardRemove {
		return fmt.Errorf("transition/reset requires hard removal")
	}
	if request.RequestedDisposition == DispositionSoftRetire && !request.Capabilities.SupportsSoftRetire {
		return fmt.Errorf("entity does not support soft retirement")
	}
	if request.RequestedDisposition == DispositionHardRemove && !request.Capabilities.SupportsHardRemove {
		return fmt.Errorf("entity does not support hard removal")
	}
	if request.Capabilities.RequiresExplicitCleanup && request.RequestedDisposition != DispositionSoftRetire {
		return fmt.Errorf("entity requires explicit cleanup through soft retirement")
	}
	if request.Capabilities.RequiresDestruction && request.RequestedDisposition != DispositionHardRemove {
		return fmt.Errorf("entity requires destruction through hard removal")
	}
	return nil
}

func ValidateOrigin(metadata OriginMetadata) error {
	if metadata.ProfileID == "" {
		return fmt.Errorf("encounter origin requires a profile ID")
	}
	if metadata.SpawnType == "" {
		return fmt.Errorf("encounter origin requires a spawn type")
	}
	if metadata.LifecyclePolicyID == "" {
		return fmt.Errorf("encounter origin requires a lifecycle policy ID")
	}
	if metadata.Priority < 0 {
		return fmt.Errorf("encounter origin priority cannot be negative")
	}
	if metadata.WeightedPopulationCost <= 0 {
		return fmt.Errorf("encounter origin weighted population cost must be positive")
	}
	return nil
}

func ValidateCapabilities(capabilities EntityCapabilities) error {
	if !capabilities.SupportsSoftRetire && !capabilities.SupportsHardRemove {
		return fmt.Errorf("entity must support soft retirement or hard removal")
	}
	if capabilities.RequiresExplicitCleanup && !capabilities.SupportsSoftRetire {
		return fmt.Errorf("explicit cleanup requires soft retirement support")
	}
	if capabilities.RequiresDestruction && !capabilities.SupportsHardRemove {
		return fmt.Errorf("destruction requires hard removal support")
	}
	if capabilities.RequiresExplicitCleanup && capabilities.RequiresDestruction {
		return fmt.Errorf("entity cannot require both explicit cleanup and destruction")
	}
	return nil
}

func isSupportedTrigger(trigger Trigger) bool {
	switch trigger {
	case TriggerLifetimeExpiry,
		TriggerOutsideAllRelevantPlayers,
		TriggerAllowedRegionExit,
		TriggerPopulationPressure,
		TriggerProfilePhaseCleanup,
		TriggerScriptedCleanup,
		TriggerTransitionReset:
		return true
	default:
		return false
	}
}

func isSupportedDisposition(disposition Disposition) bool {
	return disposition == DispositionSoftRetire || disposition == DispositionHardRemove
}
