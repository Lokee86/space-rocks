package awards

import (
	"fmt"
	"sort"
)

const (
	DefaultComboIncrease      = 0.25
	DefaultComboWindowSeconds = 0.75
)

type ComboState struct {
	Multiplier float64
	Tier       int
	ExpiresAt  float64
	Active     bool
}

type ComboOutcome struct {
	AppliedMultiplier float64
	State             ComboState
}

func (runtime *Runtime) ApplyCombo(owner Owner, now float64) (ComboOutcome, error) {
	if err := owner.Validate(); err != nil {
		return ComboOutcome{}, err
	}
	state := runtime.currentCombo(owner, now)
	applied := state.Multiplier
	state.Multiplier += DefaultComboIncrease
	state.Tier++
	state.ExpiresAt = now + DefaultComboWindowSeconds
	state.Active = true
	runtime.combos[owner] = state
	return ComboOutcome{AppliedMultiplier: applied, State: state}, nil
}

func (runtime *Runtime) Combo(owner Owner, now float64) ComboState {
	state := runtime.currentCombo(owner, now)
	if !state.Active {
		delete(runtime.combos, owner)
	}
	return state
}

func (runtime *Runtime) IncrementStreak(owner Owner, name string) (int, error) {
	if err := owner.Validate(); err != nil {
		return 0, err
	}
	if name == "" {
		return 0, fmt.Errorf("streak name is required")
	}
	if runtime.streaks[owner] == nil {
		runtime.streaks[owner] = make(map[string]int)
	}
	runtime.streaks[owner][name]++
	return runtime.streaks[owner][name], nil
}

func (runtime *Runtime) ResetProgress(owner Owner) {
	delete(runtime.combos, owner)
	delete(runtime.streaks, owner)
}

func (runtime *Runtime) ComboSnapshots(now float64) []ComboSnapshot {
	result := make([]ComboSnapshot, 0, len(runtime.combos))
	for owner := range runtime.combos {
		state := runtime.Combo(owner, now)
		if state.Active {
			result = append(result, ComboSnapshot{Owner: owner, State: state})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Owner.Scope != result[j].Owner.Scope {
			return result[i].Owner.Scope < result[j].Owner.Scope
		}
		return result[i].Owner.ID < result[j].Owner.ID
	})
	return result
}

func (runtime *Runtime) StreakSnapshot() []StreakSnapshot {
	result := make([]StreakSnapshot, 0)
	for owner, streaks := range runtime.streaks {
		for name, count := range streaks {
			result = append(result, StreakSnapshot{Owner: owner, Name: name, Count: count})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Owner.Scope != result[j].Owner.Scope {
			return result[i].Owner.Scope < result[j].Owner.Scope
		}
		if result[i].Owner.ID != result[j].Owner.ID {
			return result[i].Owner.ID < result[j].Owner.ID
		}
		return result[i].Name < result[j].Name
	})
	return result
}

func (runtime *Runtime) currentCombo(owner Owner, now float64) ComboState {
	state, ok := runtime.combos[owner]
	if !ok || !state.Active || now > state.ExpiresAt {
		return ComboState{Multiplier: 1.0}
	}
	if state.Multiplier < 1.0 {
		state.Multiplier = 1.0
	}
	return state
}
