package objectives

import "testing"

func TestSequenceSetAndMaintainConditions(t *testing.T) {
	sequence := Condition{Kind: ConditionSequence, Sequence: []string{"a", "b"}}
	sequenceState := newConditionState(sequence)
	applyFacts(sequence, &sequenceState, []Fact{{Key: "b"}, {Key: "a"}, {Key: "b"}})
	if !conditionSatisfied(sequence, sequenceState) {
		t.Fatalf("sequence state = %#v", sequenceState)
	}

	set := Condition{Kind: ConditionSet, FactKey: "collect", RequiredMembers: []string{"red", "blue"}}
	setState := newConditionState(set)
	applyFacts(set, &setState, []Fact{
		{Key: "collect", Operation: FactAddMember, Member: "red"},
		{Key: "collect", Operation: FactAddMember, Member: "blue"},
	})
	if !conditionSatisfied(set, setState) {
		t.Fatalf("set state = %#v", setState)
	}

	maintain := Condition{Kind: ConditionMaintain, FactKey: "safe", Expected: true, Target: 3, AllowReset: true}
	maintainState := newConditionState(maintain)
	applyFacts(maintain, &maintainState, []Fact{{Key: "safe", Boolean: Bool(true)}})
	advanceMaintain(maintain, &maintainState, 3)
	if !conditionSatisfied(maintain, maintainState) {
		t.Fatalf("maintain state = %#v", maintainState)
	}
	applyFacts(maintain, &maintainState, []Fact{{Key: "safe", Boolean: Bool(false)}})
	if maintainState.Progress != 0 {
		t.Fatalf("maintain progress = %v, want reset", maintainState.Progress)
	}
}

func TestAttributionFilterConsumesOnlyAllowedFacts(t *testing.T) {
	condition := Condition{
		Kind:               ConditionNumeric,
		FactKey:            "damage",
		Target:             10,
		AllowedAttribution: []AttributionKind{AttributionInEncounter},
	}
	state := newConditionState(condition)
	applyFacts(condition, &state, []Fact{
		{Key: "damage", Number: 10, Attribution: Attribution{Kind: AttributionOneHit}},
		{Key: "damage", Number: 10, Attribution: Attribution{Kind: AttributionInEncounter, PlayerID: "player-1"}},
	})
	if state.Progress != 10 || state.Attribution.PlayerID != "player-1" {
		t.Fatalf("state = %#v", state)
	}
}
