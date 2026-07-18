package awards

import (
	"fmt"
	"sort"
)

type Runtime struct {
	customCounters map[CounterID]struct{}
	visibility     map[CounterID]Visibility
	values         map[Owner]map[CounterID]float64
	processed      map[string]struct{}
	combos         map[Owner]ComboState
	streaks        map[Owner]map[string]int
	contributions  map[string][]Contribution
}

func NewRuntime() *Runtime {
	return &Runtime{
		customCounters: make(map[CounterID]struct{}),
		visibility: map[CounterID]Visibility{
			CounterScore: VisibilityScoreboard, CounterKills: VisibilityScoreboard,
			CounterAssists: VisibilityScoreboard, CounterDeaths: VisibilityScoreboard,
			CounterDamageDealt: VisibilityResultsOnly, CounterDamageTaken: VisibilityResultsOnly,
			CounterObjectiveProgress: VisibilityHUD, CounterResourcesCollected: VisibilityHUD,
			CounterCompletionTime: VisibilityResultsOnly,
		},
		values:        make(map[Owner]map[CounterID]float64),
		processed:     make(map[string]struct{}),
		combos:        make(map[Owner]ComboState),
		streaks:       make(map[Owner]map[string]int),
		contributions: make(map[string][]Contribution),
	}
}

func (runtime *Runtime) RegisterCustomCounter(id CounterID) error {
	if err := ValidateCustomCounterID(id); err != nil {
		return err
	}
	runtime.customCounters[id] = struct{}{}
	runtime.visibility[id] = VisibilityHidden
	return nil
}

func (runtime *Runtime) RegisterCustomCounterWithVisibility(id CounterID, visibility Visibility) error {
	if err := runtime.RegisterCustomCounter(id); err != nil {
		return err
	}
	return runtime.SetCounterVisibility(id, visibility)
}

func (runtime *Runtime) SetCounterVisibility(id CounterID, visibility Visibility) error {
	if !runtime.CounterRegistered(id) {
		return fmt.Errorf("counter %q is not registered", id)
	}
	switch visibility {
	case VisibilityHidden, VisibilityHUD, VisibilityScoreboard, VisibilityResultsOnly,
		VisibilityTeamOnly, VisibilityPlayerPrivate, VisibilitySpectatorVisible:
		runtime.visibility[id] = visibility
		return nil
	default:
		return fmt.Errorf("invalid visibility %q", visibility)
	}
}

func (runtime *Runtime) CounterVisibility(id CounterID) (Visibility, bool) {
	visibility, ok := runtime.visibility[id]
	return visibility, ok
}

func (runtime *Runtime) CounterRegistered(id CounterID) bool {
	if IsFixedCounter(id) {
		return true
	}
	_, ok := runtime.customCounters[id]
	return ok
}

func (runtime *Runtime) ApplyEvent(eventID string, mutations []Mutation) (EventResult, error) {
	if eventID == "" {
		return EventResult{}, fmt.Errorf("award event id is required")
	}
	if _, duplicate := runtime.processed[eventID]; duplicate {
		return EventResult{EventID: eventID, Duplicate: true}, nil
	}
	for _, mutation := range mutations {
		if err := runtime.validateMutation(mutation); err != nil {
			return EventResult{}, err
		}
	}

	result := EventResult{EventID: eventID, Applied: true, Changes: make([]Change, 0, len(mutations))}
	for _, mutation := range mutations {
		ownerValues := runtime.values[mutation.Owner]
		if ownerValues == nil {
			ownerValues = make(map[CounterID]float64)
			runtime.values[mutation.Owner] = ownerValues
		}
		before := ownerValues[mutation.CounterID]
		after := applyMutation(before, mutation)
		ownerValues[mutation.CounterID] = after
		result.Changes = append(result.Changes, Change{
			Owner: mutation.Owner, CounterID: mutation.CounterID, Operation: mutation.Operation,
			Before: before, After: after, Delta: after - before, Source: mutation.Source,
		})
	}
	runtime.processed[eventID] = struct{}{}
	return result, nil
}

func (runtime *Runtime) Reset() {
	runtime.values = make(map[Owner]map[CounterID]float64)
	runtime.processed = make(map[string]struct{})
	runtime.combos = make(map[Owner]ComboState)
	runtime.streaks = make(map[Owner]map[string]int)
	runtime.contributions = make(map[string][]Contribution)
}

func (runtime *Runtime) RemoveOwner(owner Owner) {
	delete(runtime.values, owner)
	delete(runtime.combos, owner)
	delete(runtime.streaks, owner)
	if owner.Scope != ScopePlayer {
		return
	}
	for targetID, contributions := range runtime.contributions {
		kept := contributions[:0]
		for _, contribution := range contributions {
			if contribution.PlayerID != owner.ID {
				kept = append(kept, contribution)
			}
		}
		if len(kept) == 0 {
			delete(runtime.contributions, targetID)
		} else {
			runtime.contributions[targetID] = kept
		}
	}
}

func (runtime *Runtime) EventProcessed(eventID string) bool {
	_, ok := runtime.processed[eventID]
	return ok
}

func (runtime *Runtime) Counter(owner Owner, id CounterID) (float64, bool) {
	ownerValues, ok := runtime.values[owner]
	if !ok {
		return 0, false
	}
	value, ok := ownerValues[id]
	return value, ok
}

func (runtime *Runtime) Snapshot() []CounterSnapshot {
	snapshot := make([]CounterSnapshot, 0)
	for owner, counters := range runtime.values {
		for id, value := range counters {
			snapshot = append(snapshot, CounterSnapshot{Owner: owner, CounterID: id, Value: value, Visibility: runtime.visibility[id]})
		}
	}
	sort.Slice(snapshot, func(i, j int) bool {
		if snapshot[i].Owner.Scope != snapshot[j].Owner.Scope {
			return snapshot[i].Owner.Scope < snapshot[j].Owner.Scope
		}
		if snapshot[i].Owner.ID != snapshot[j].Owner.ID {
			return snapshot[i].Owner.ID < snapshot[j].Owner.ID
		}
		return snapshot[i].CounterID < snapshot[j].CounterID
	})
	return snapshot
}

func (runtime *Runtime) DerivedTeamTotals(memberships map[string]string, ids []CounterID) []CounterSnapshot {
	totals := make(map[Owner]map[CounterID]float64)
	for playerID, teamID := range memberships {
		if teamID == "" {
			continue
		}
		playerOwner := Owner{Scope: ScopePlayer, ID: playerID}
		teamOwner := Owner{Scope: ScopeTeam, ID: teamID}
		for _, id := range ids {
			value, _ := runtime.Counter(playerOwner, id)
			if totals[teamOwner] == nil {
				totals[teamOwner] = make(map[CounterID]float64)
			}
			totals[teamOwner][id] += value
		}
	}
	result := make([]CounterSnapshot, 0)
	for owner, counters := range totals {
		for id, value := range counters {
			result = append(result, CounterSnapshot{Owner: owner, CounterID: id, Value: value, Visibility: runtime.visibility[id]})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Owner.ID != result[j].Owner.ID {
			return result[i].Owner.ID < result[j].Owner.ID
		}
		return result[i].CounterID < result[j].CounterID
	})
	return result
}

func (runtime *Runtime) validateMutation(mutation Mutation) error {
	if err := mutation.Owner.Validate(); err != nil {
		return err
	}
	if !runtime.CounterRegistered(mutation.CounterID) {
		return fmt.Errorf("counter %q is not registered", mutation.CounterID)
	}
	switch mutation.Operation {
	case MutationIncrement, MutationDecrement, MutationSet, MutationMinimum, MutationMaximum, MutationTimedAccumulate:
		return nil
	default:
		return fmt.Errorf("invalid mutation operation %q", mutation.Operation)
	}
}

func applyMutation(before float64, mutation Mutation) float64 {
	switch mutation.Operation {
	case MutationIncrement, MutationTimedAccumulate:
		return before + mutation.Value
	case MutationDecrement:
		return before - mutation.Value
	case MutationSet:
		return mutation.Value
	case MutationMinimum:
		return max(before, mutation.Value)
	case MutationMaximum:
		return min(before, mutation.Value)
	default:
		return before
	}
}
