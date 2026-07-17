package teams

import (
	"reflect"
	"testing"
)

func TestResolveAssignmentsFFA(t *testing.T) {
	assignments, err := ResolveAssignments(Config{Structure: StructureFFA}, []string{"member-b", "member-a"}, nil)
	if err != nil {
		t.Fatalf("resolve FFA assignments: %v", err)
	}
	want := Assignments{"member-a": NoTeam, "member-b": NoTeam}
	if !reflect.DeepEqual(assignments, want) {
		t.Fatalf("assignments = %+v, want %+v", assignments, want)
	}
}

func TestResolveAssignmentsCoOp(t *testing.T) {
	assignments, err := ResolveAssignments(Config{Structure: StructureCoOp}, []string{"member-b", "member-a"}, nil)
	if err != nil {
		t.Fatalf("resolve co-op assignments: %v", err)
	}
	want := Assignments{"member-a": Team1, "member-b": Team1}
	if !reflect.DeepEqual(assignments, want) {
		t.Fatalf("assignments = %+v, want %+v", assignments, want)
	}
}

func TestResolveAssignmentsCustom(t *testing.T) {
	requested := Assignments{"member-a": Team2, "member-b": Team1}
	assignments, err := ResolveAssignments(Config{
		Structure:      StructureCustom,
		AssignmentMode: AssignmentOwnerAssigned,
	}, []string{"member-b", "member-a"}, requested)
	if err != nil {
		t.Fatalf("resolve custom assignments: %v", err)
	}
	if !reflect.DeepEqual(assignments, requested) {
		t.Fatalf("assignments = %+v, want %+v", assignments, requested)
	}
}

func TestResolveAutoBalancedIsCanonicalAndDeterministic(t *testing.T) {
	config := Config{Structure: StructureAutoBalanced, AutoTeamCount: 2}
	first, err := ResolveAssignments(config, []string{"member-d", "member-b", "member-c", "member-a"}, nil)
	if err != nil {
		t.Fatalf("resolve first auto-balanced roster: %v", err)
	}
	second, err := ResolveAssignments(config, []string{"member-a", "member-c", "member-b", "member-d"}, nil)
	if err != nil {
		t.Fatalf("resolve second auto-balanced roster: %v", err)
	}
	want := Assignments{
		"member-a": Team1,
		"member-b": Team2,
		"member-c": Team1,
		"member-d": Team2,
	}
	if !reflect.DeepEqual(first, want) || !reflect.DeepEqual(second, want) {
		t.Fatalf("canonical assignments = %+v/%+v, want %+v", first, second, want)
	}
}

func TestValidateConfigRejectsInvalidConfigurations(t *testing.T) {
	configs := []Config{
		{Structure: "unknown"},
		{Structure: StructureFFA, AssignmentMode: AssignmentPlayerSelected},
		{Structure: StructureCoOp, AutoTeamCount: 2},
		{Structure: StructureCustom, AssignmentMode: "unknown"},
		{Structure: StructureAutoBalanced, AutoTeamCount: 1},
		{Structure: StructureAutoBalanced, AutoTeamCount: len(OrderedIDs()) + 1},
	}
	for _, config := range configs {
		if err := ValidateConfig(config); err == nil {
			t.Fatalf("expected config %+v to be rejected", config)
		}
	}
}

func TestCanonicalRosterRejectsBlankAndDuplicateParticipants(t *testing.T) {
	for _, participants := range [][]string{
		{"member-a", ""},
		{"member-a", "member-a"},
	} {
		if _, err := CanonicalRoster(participants); err == nil {
			t.Fatalf("expected roster %q to be rejected", participants)
		}
	}
}
