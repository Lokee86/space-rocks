package objectives

import "math"

type conditionState struct {
	Progress        float64
	Satisfied       bool
	SequenceIndex   int
	Members         map[string]struct{}
	MaintainEnabled bool
	Attribution     Attribution
}

func newConditionState(condition Condition) conditionState {
	state := conditionState{}
	if condition.Kind == ConditionSet {
		state.Members = make(map[string]struct{})
	}
	return state
}

func applyFacts(condition Condition, state *conditionState, facts []Fact) (bool, string) {
	before := *state
	before.Members = cloneMembers(state.Members)
	lastKey := ""
	for _, fact := range facts {
		if !conditionAcceptsFact(condition, fact) {
			continue
		}
		lastKey = fact.Key
		state.Attribution = fact.Attribution
		switch condition.Kind {
		case ConditionBoolean:
			if fact.Boolean != nil {
				state.Satisfied = *fact.Boolean == condition.Expected
			}
		case ConditionNumeric:
			applyNumericFact(condition, state, fact)
		case ConditionSet:
			applySetFact(state, fact)
		case ConditionSequence:
			applySequenceFact(condition, state, fact)
		case ConditionMaintain:
			if fact.Boolean != nil {
				state.MaintainEnabled = *fact.Boolean == condition.Expected
				if !state.MaintainEnabled && condition.AllowReset {
					state.Progress = 0
				}
			}
		}
	}
	return !conditionStatesEqual(before, *state), lastKey
}

func applyNumericFact(condition Condition, state *conditionState, fact Fact) {
	switch fact.Operation {
	case FactReset:
		if condition.AllowReset {
			state.Progress = 0
		}
	case FactSet:
		if condition.AllowDecrease || fact.Number >= state.Progress {
			state.Progress = fact.Number
		}
	case FactIncrement, "":
		if fact.Number >= 0 || condition.AllowDecrease {
			state.Progress += fact.Number
		}
	}
	state.Progress = normalizedProgress(condition, state.Progress)
}

func applySetFact(state *conditionState, fact Fact) {
	if fact.Member == "" {
		return
	}
	if state.Members == nil {
		state.Members = make(map[string]struct{})
	}
	switch fact.Operation {
	case FactRemoveMember:
		delete(state.Members, fact.Member)
	case FactAddMember, "":
		state.Members[fact.Member] = struct{}{}
	}
	state.Progress = float64(len(state.Members))
}

func applySequenceFact(condition Condition, state *conditionState, fact Fact) {
	if fact.Operation == FactReset && condition.AllowReset {
		state.SequenceIndex = 0
		state.Progress = 0
		return
	}
	if state.SequenceIndex >= len(condition.Sequence) || fact.Key != condition.Sequence[state.SequenceIndex] {
		return
	}
	state.SequenceIndex++
	state.Progress = float64(state.SequenceIndex)
}

func advanceMaintain(condition Condition, state *conditionState, delta float64) bool {
	if condition.Kind != ConditionMaintain || !state.MaintainEnabled || delta <= 0 {
		return false
	}
	before := state.Progress
	state.Progress = normalizedProgress(condition, state.Progress+delta)
	return state.Progress != before
}

func conditionSatisfied(condition Condition, state conditionState) bool {
	switch condition.Kind {
	case ConditionManual:
		return false
	case ConditionBoolean:
		return state.Satisfied
	case ConditionNumeric, ConditionMaintain:
		return state.Progress >= condition.Target
	case ConditionSet:
		for _, member := range condition.RequiredMembers {
			if _, ok := state.Members[member]; !ok {
				return false
			}
		}
		return true
	case ConditionSequence:
		return state.SequenceIndex >= len(condition.Sequence)
	default:
		return false
	}
}

func conditionTarget(condition Condition) float64 {
	switch condition.Kind {
	case ConditionBoolean:
		return 1
	case ConditionNumeric, ConditionMaintain:
		return condition.Target
	case ConditionSet:
		return float64(len(condition.RequiredMembers))
	case ConditionSequence:
		return float64(len(condition.Sequence))
	default:
		return 0
	}
}

func normalizedProgress(condition Condition, value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	if value < 0 {
		value = 0
	}
	if condition.Overflow == OverflowClamp && condition.Target > 0 && value > condition.Target {
		return condition.Target
	}
	return value
}

func conditionAcceptsFact(condition Condition, fact Fact) bool {
	if condition.Kind == ConditionSequence {
		return attributionAllowed(condition, fact.Attribution.Kind)
	}
	if condition.FactKey == "" || condition.FactKey != fact.Key {
		return false
	}
	return attributionAllowed(condition, fact.Attribution.Kind)
}

func attributionAllowed(condition Condition, kind AttributionKind) bool {
	if len(condition.AllowedAttribution) == 0 {
		return true
	}
	for _, allowed := range condition.AllowedAttribution {
		if allowed == kind {
			return true
		}
	}
	return false
}
