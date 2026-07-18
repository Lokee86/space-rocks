package encounterlifecycle

import (
	"fmt"
	"math"
	"sort"
)

// Registration is the immutable contract captured when an encounter entity is
// admitted to the lifecycle runtime.
type Registration struct {
	Origin       OriginMetadata
	Policy       Policy
	Capabilities EntityCapabilities
}

func (registration Registration) Validate() error {
	if err := ValidateOrigin(registration.Origin); err != nil {
		return err
	}
	if err := ValidateCapabilities(registration.Capabilities); err != nil {
		return err
	}
	if err := registration.Policy.Validate(); err != nil {
		return err
	}

	configured := []struct {
		trigger Trigger
		policy  TriggerPolicy
	}{
		{TriggerLifetimeExpiry, registration.Policy.LifetimeExpiry},
		{TriggerOutsideAllRelevantPlayers, registration.Policy.OutsideAllRelevantPlayers},
		{TriggerAllowedRegionExit, registration.Policy.AllowedRegionExit},
		{TriggerPopulationPressure, registration.Policy.PopulationPressure},
		{TriggerProfilePhaseCleanup, registration.Policy.ProfilePhaseCleanup},
		{TriggerScriptedCleanup, registration.Policy.ScriptedCleanup},
		{TriggerTransitionReset, registration.Policy.TransitionReset},
	}
	for _, candidate := range configured {
		if !candidate.policy.Enabled {
			continue
		}
		if _, err := Decide(DecisionRequest{
			Origin:               registration.Origin,
			Trigger:              candidate.trigger,
			Capabilities:         registration.Capabilities,
			RequestedDisposition: candidate.policy.Disposition,
		}); err != nil {
			return fmt.Errorf("invalid %s registration policy: %w", candidate.trigger, err)
		}
	}
	return nil
}

// RetirementState identifies whether an entry remains active or has entered
// the authoritative cleanup handoff.
type RetirementState string

const (
	RetirementStateActive RetirementState = "active"
	RetirementStateBegun  RetirementState = "retirement_begun"
)

// Entry is an immutable runtime snapshot. Returning it by value prevents
// callers from mutating registry state or retaining internal pointers.
type Entry struct {
	EntityID               string
	Registration           Registration
	ElapsedLifetimeSeconds float64
	RetirementState        RetirementState
	Retirement             EvaluationResult
}

// Runtime owns registered encounter lifecycle state and origin-profile
// weighted-population accounting.
type Runtime struct {
	entries       map[string]Entry
	profileTotals map[ProfileID]WeightedPopulationCost
}

func NewRuntime() *Runtime {
	return &Runtime{
		entries:       make(map[string]Entry),
		profileTotals: make(map[ProfileID]WeightedPopulationCost),
	}
}

func (runtime *Runtime) Register(entityID string, registration Registration) error {
	if entityID == "" {
		return fmt.Errorf("encounter entity ID is required")
	}
	if err := registration.Validate(); err != nil {
		return fmt.Errorf("invalid encounter registration: %w", err)
	}
	if _, exists := runtime.entries[entityID]; exists {
		return fmt.Errorf("encounter entity %q is already registered", entityID)
	}
	if runtime.entries == nil {
		runtime.entries = make(map[string]Entry)
		runtime.profileTotals = make(map[ProfileID]WeightedPopulationCost)
	}

	runtime.entries[entityID] = Entry{
		EntityID:        entityID,
		Registration:    registration,
		RetirementState: RetirementStateActive,
	}
	runtime.profileTotals[registration.Origin.ProfileID] += registration.Origin.WeightedPopulationCost
	return nil
}

func (runtime *Runtime) Advance(delta float64, simulationPaused bool) error {
	if delta < 0 || math.IsNaN(delta) || math.IsInf(delta, 0) {
		return fmt.Errorf("lifecycle advance delta must be finite and non-negative")
	}
	if simulationPaused {
		return nil
	}
	for entityID, entry := range runtime.entries {
		if entry.RetirementState != RetirementStateActive {
			continue
		}
		entry.ElapsedLifetimeSeconds += delta
		runtime.entries[entityID] = entry
	}
	return nil
}

func (runtime *Runtime) Snapshot(entityID string) (Entry, bool) {
	entry, ok := runtime.entries[entityID]
	return entry, ok
}

func (runtime *Runtime) EntityIDs() []string {
	entityIDs := make([]string, 0, len(runtime.entries))
	for entityID := range runtime.entries {
		entityIDs = append(entityIDs, entityID)
	}
	sort.Strings(entityIDs)
	return entityIDs
}

func (runtime *Runtime) ProfileWeightedPopulationTotals() map[ProfileID]WeightedPopulationCost {
	totals := make(map[ProfileID]WeightedPopulationCost, len(runtime.profileTotals))
	for profileID, total := range runtime.profileTotals {
		totals[profileID] = total
	}
	return totals
}

func (runtime *Runtime) BeginRetirement(entityID string, result EvaluationResult) error {
	entry, ok := runtime.entries[entityID]
	if !ok {
		return fmt.Errorf("encounter entity %q is not registered", entityID)
	}
	if entry.RetirementState != RetirementStateActive {
		return fmt.Errorf("encounter entity %q has already begun retirement", entityID)
	}
	decision, err := Decide(DecisionRequest{
		Origin:               entry.Registration.Origin,
		Trigger:              result.Trigger,
		Capabilities:         entry.Registration.Capabilities,
		RequestedDisposition: result.Disposition,
	})
	if err != nil {
		return fmt.Errorf("invalid retirement for encounter entity %q: %w", entityID, err)
	}
	entry.RetirementState = RetirementStateBegun
	entry.Retirement = EvaluationResult{Trigger: result.Trigger, Disposition: decision.Disposition}
	runtime.entries[entityID] = entry
	return nil
}

func (runtime *Runtime) Remove(entityID string) (Entry, bool) {
	entry, ok := runtime.entries[entityID]
	if !ok {
		return Entry{}, false
	}
	delete(runtime.entries, entityID)

	profileID := entry.Registration.Origin.ProfileID
	total := runtime.profileTotals[profileID] - entry.Registration.Origin.WeightedPopulationCost
	if total <= 0 {
		delete(runtime.profileTotals, profileID)
	} else {
		runtime.profileTotals[profileID] = total
	}
	return entry, true
}
