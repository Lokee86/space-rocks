package encounterspawn

import (
	"fmt"
	"math"
	"sort"
)

type ActivationState string

const (
	StateConfigured  ActivationState = "configured"
	StateActive      ActivationState = "active"
	StateDeactivated ActivationState = "deactivated"
)

type Opportunity struct {
	ProfileID ProfileID
	BatchSize int
	Priority  Priority
}

type Snapshot struct {
	Config         Config
	State          ActivationState
	ElapsedSeconds float64
	RuntimeStopped bool
}

type profileState struct {
	config  Config
	state   ActivationState
	elapsed float64
}

type Runtime struct {
	profiles map[ProfileID]profileState
	stopped  bool
}

func NewRuntime() *Runtime {
	return &Runtime{profiles: make(map[ProfileID]profileState)}
}

func (runtime *Runtime) Configure(config Config) error {
	if err := config.Validate(); err != nil {
		return err
	}
	if runtime.profiles == nil {
		runtime.profiles = make(map[ProfileID]profileState)
	}
	if _, exists := runtime.profiles[config.ID]; exists {
		return fmt.Errorf("encounter spawn profile %q is already configured", config.ID)
	}
	state := StateConfigured
	if config.InitiallyActive {
		state = StateActive
	}
	runtime.profiles[config.ID] = profileState{config: config.Clone(), state: state}
	return nil
}

func (runtime *Runtime) Activate(profileID ProfileID) error {
	profile, ok := runtime.profiles[profileID]
	if !ok {
		return fmt.Errorf("encounter spawn profile %q is not configured", profileID)
	}
	profile.state = StateActive
	runtime.profiles[profileID] = profile
	return nil
}

func (runtime *Runtime) Deactivate(profileID ProfileID) error {
	profile, ok := runtime.profiles[profileID]
	if !ok {
		return fmt.Errorf("encounter spawn profile %q is not configured", profileID)
	}
	profile.state = StateDeactivated
	runtime.profiles[profileID] = profile
	return nil
}

func (runtime *Runtime) Stop() {
	runtime.stopped = true
}

func (runtime *Runtime) Resume() {
	runtime.stopped = false
}

func (runtime *Runtime) ResetProgress(profileID ProfileID) error {
	profile, ok := runtime.profiles[profileID]
	if !ok {
		return fmt.Errorf("encounter spawn profile %q is not configured", profileID)
	}
	profile.elapsed = 0
	runtime.profiles[profileID] = profile
	return nil
}

func (runtime *Runtime) Step(delta float64, simulationPaused bool) ([]Opportunity, error) {
	if delta < 0 || math.IsNaN(delta) || math.IsInf(delta, 0) {
		return nil, fmt.Errorf("encounter spawn step delta must be finite and non-negative")
	}
	if simulationPaused || runtime.stopped {
		return nil, nil
	}

	profileIDs := runtime.priorityOrderedProfileIDs()
	opportunities := make([]Opportunity, 0)
	for _, profileID := range profileIDs {
		profile := runtime.profiles[profileID]
		if profile.state != StateActive || profile.config.ScheduleKind != ScheduleContinuous {
			continue
		}
		profile.elapsed += delta
		due := int(math.Floor(profile.elapsed / profile.config.IntervalSeconds))
		if due > 0 {
			profile.elapsed -= float64(due) * profile.config.IntervalSeconds
			for range due {
				opportunities = append(opportunities, Opportunity{
					ProfileID: profileID,
					BatchSize: profile.config.BatchSize,
					Priority:  profile.config.Priority,
				})
			}
		}
		runtime.profiles[profileID] = profile
	}
	return opportunities, nil
}

func (runtime *Runtime) Snapshot(profileID ProfileID) (Snapshot, bool) {
	profile, ok := runtime.profiles[profileID]
	if !ok {
		return Snapshot{}, false
	}
	return Snapshot{
		Config:         profile.config.Clone(),
		State:          profile.state,
		ElapsedSeconds: profile.elapsed,
		RuntimeStopped: runtime.stopped,
	}, true
}

func (runtime *Runtime) ProfileIDs() []ProfileID {
	profileIDs := make([]ProfileID, 0, len(runtime.profiles))
	for profileID := range runtime.profiles {
		profileIDs = append(profileIDs, profileID)
	}
	sort.Slice(profileIDs, func(i, j int) bool { return profileIDs[i] < profileIDs[j] })
	return profileIDs
}

func (runtime *Runtime) priorityOrderedProfileIDs() []ProfileID {
	profileIDs := runtime.ProfileIDs()
	sort.SliceStable(profileIDs, func(i, j int) bool {
		left := runtime.profiles[profileIDs[i]].config.Priority
		right := runtime.profiles[profileIDs[j]].config.Priority
		return left > right
	})
	return profileIDs
}
