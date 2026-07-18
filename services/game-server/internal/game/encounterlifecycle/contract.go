package encounterlifecycle

// ProfileID identifies the encounter profile that originated an entity.
type ProfileID string

// SpawnType identifies the kind of encounter spawn that produced an entity.
type SpawnType string

// LifecyclePolicyID identifies the post-spawn lifecycle policy for an entity.
type LifecyclePolicyID string

// Priority orders otherwise eligible lifecycle candidates within a profile.
type Priority int

// WeightedPopulationCost is the amount charged to the originating profile.
type WeightedPopulationCost int

// PhaseID identifies an optional encounter phase association.
type PhaseID string

// OriginMetadata is the immutable origin/accounting metadata carried by an
// eligible non-player encounter entity.
type OriginMetadata struct {
	ProfileID              ProfileID
	SpawnType              SpawnType
	LifecyclePolicyID      LifecyclePolicyID
	Priority               Priority
	WeightedPopulationCost WeightedPopulationCost
	PhaseID                PhaseID
}

func (metadata OriginMetadata) HasPhase() bool {
	return metadata.PhaseID != ""
}

// Trigger identifies an authoritative reason to evaluate encounter retirement.
type Trigger string

const (
	TriggerLifetimeExpiry            Trigger = "lifetime_expiry"
	TriggerOutsideAllRelevantPlayers Trigger = "outside_all_relevant_players"
	TriggerAllowedRegionExit         Trigger = "allowed_region_exit"
	TriggerPopulationPressure        Trigger = "population_pressure"
	TriggerProfilePhaseCleanup       Trigger = "profile_phase_cleanup"
	TriggerScriptedCleanup           Trigger = "scripted_cleanup"
	TriggerTransitionReset           Trigger = "transition_reset"
)

// Disposition is the cleanup handoff selected for a lifecycle request.
type Disposition string

const (
	DispositionSoftRetire Disposition = "soft_retire"
	DispositionHardRemove Disposition = "hard_remove"
)

// EntityCapabilities declares the cleanup paths an entity type can honor.
type EntityCapabilities struct {
	SupportsSoftRetire      bool
	SupportsHardRemove      bool
	RequiresExplicitCleanup bool
	RequiresDestruction     bool
}
