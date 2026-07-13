package health

import (
	"sync"
	"time"
)

type State struct {
	mu                                                                   sync.RWMutex
	serviceInstanceID, buildVersion, environment                         string
	startedAt                                                            time.Time
	ready, stopping                                                      bool
	batchesReceived, eventsAccepted, eventsRejected, eventsRedacted      uint64
	duplicatesSuppressed, storageFailures, rotations, retentionDeletions uint64
	queryFailures, diagnosticBundlesCreated                              uint64
}

type Snapshot struct {
	ServiceInstanceID        string    `json:"service_instance_id"`
	BuildVersion             string    `json:"build_version"`
	Environment              string    `json:"environment"`
	StartedAt                time.Time `json:"started_at"`
	Ready                    bool      `json:"ready"`
	Stopping                 bool      `json:"stopping"`
	BatchesReceived          uint64    `json:"batches_received"`
	EventsAccepted           uint64    `json:"events_accepted"`
	EventsRejected           uint64    `json:"events_rejected"`
	EventsRedacted           uint64    `json:"events_redacted"`
	DuplicatesSuppressed     uint64    `json:"duplicates_suppressed"`
	StorageFailures          uint64    `json:"storage_failures"`
	Rotations                uint64    `json:"rotations"`
	RetentionDeletions       uint64    `json:"retention_deletions"`
	QueryFailures            uint64    `json:"query_failures"`
	DiagnosticBundlesCreated uint64    `json:"diagnostic_bundles_created"`
}

func NewState(serviceInstanceID, buildVersion, environment string, startedAt time.Time) *State {
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	return &State{serviceInstanceID: serviceInstanceID, buildVersion: buildVersion, environment: environment, startedAt: startedAt}
}

func (s *State) MarkReady()    { s.mu.Lock(); s.ready = true; s.mu.Unlock() }
func (s *State) MarkStopping() { s.mu.Lock(); s.stopping = true; s.ready = false; s.mu.Unlock() }
func (s *State) Snapshot() Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Snapshot{ServiceInstanceID: s.serviceInstanceID, BuildVersion: s.buildVersion, Environment: s.environment, StartedAt: s.startedAt, Ready: s.ready, Stopping: s.stopping, BatchesReceived: s.batchesReceived, EventsAccepted: s.eventsAccepted, EventsRejected: s.eventsRejected, EventsRedacted: s.eventsRedacted, DuplicatesSuppressed: s.duplicatesSuppressed, StorageFailures: s.storageFailures, Rotations: s.rotations, RetentionDeletions: s.retentionDeletions, QueryFailures: s.queryFailures, DiagnosticBundlesCreated: s.diagnosticBundlesCreated}
}

func (s *State) AddBatchesReceived(amount uint64) {
	s.mu.Lock()
	s.batchesReceived += amount
	s.mu.Unlock()
}
func (s *State) AddEventsAccepted(amount int) {
	s.mu.Lock()
	s.eventsAccepted += uint64(nonNegative(amount))
	s.mu.Unlock()
}
func (s *State) AddEventsRejected(amount int) {
	s.mu.Lock()
	s.eventsRejected += uint64(nonNegative(amount))
	s.mu.Unlock()
}
func (s *State) AddEventsRedacted(amount int) {
	s.mu.Lock()
	s.eventsRedacted += uint64(nonNegative(amount))
	s.mu.Unlock()
}
func (s *State) AddQueryFailures(amount uint64) {
	s.mu.Lock()
	s.queryFailures += amount
	s.mu.Unlock()
}

func nonNegative(amount int) int {
	if amount < 0 {
		return 0
	}
	return amount
}

func (s *State) IncBatchesReceived()      { s.AddBatchesReceived(1) }
func (s *State) IncEventsAccepted()       { s.AddEventsAccepted(1) }
func (s *State) IncEventsRejected()       { s.AddEventsRejected(1) }
func (s *State) IncEventsRedacted()       { s.AddEventsRedacted(1) }
func (s *State) IncDuplicatesSuppressed() { s.mu.Lock(); s.duplicatesSuppressed++; s.mu.Unlock() }
func (s *State) IncStorageFailures()      { s.mu.Lock(); s.storageFailures++; s.mu.Unlock() }
func (s *State) IncRotations()            { s.mu.Lock(); s.rotations++; s.mu.Unlock() }
func (s *State) IncRetentionDeletions()   { s.mu.Lock(); s.retentionDeletions++; s.mu.Unlock() }
func (s *State) IncQueryFailures()        { s.mu.Lock(); s.queryFailures++; s.mu.Unlock() }
func (s *State) IncDiagnosticBundlesCreated() {
	s.mu.Lock()
	s.diagnosticBundlesCreated++
	s.mu.Unlock()
}
