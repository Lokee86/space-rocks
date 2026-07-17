package teams

import (
	"reflect"
	"testing"
)

func TestResolveAssignmentsCoversAllStructures(t *testing.T) {
	participants := []string{"player-2", "player-1"}
	tests := []struct {
		name      string
		config    Config
		requested Assignments
		want      Assignments
	}{
		{
			name:   "ffa",
			config: Config{Structure: StructureFFA},
			want:   Assignments{"player-1": NoTeam, "player-2": NoTeam},
		},
		{
			name:   "co-op",
			config: Config{Structure: StructureCoOp},
			want:   Assignments{"player-1": Team1, "player-2": Team1},
		},
		{
			name:      "custom",
			config:    Config{Structure: StructureCustom, AssignmentMode: AssignmentOwnerAssigned},
			requested: Assignments{"player-1": Team2, "player-2": Team1},
			want:      Assignments{"player-1": Team2, "player-2": Team1},
		},
		{
			name:   "auto-balanced",
			config: Config{Structure: StructureAutoBalanced, AutoTeamCount: 2},
			want:   Assignments{"player-1": Team1, "player-2": Team2},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveAssignments(test.config, participants, test.requested)
			if err != nil {
				t.Fatalf("ResolveAssignments() error = %v", err)
			}
			if !reflect.DeepEqual(got, test.want) {
				t.Fatalf("ResolveAssignments() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestResolveAssignmentsRejectsRequestedAssignmentsOutsideCustom(t *testing.T) {
	requested := Assignments{"player-1": Team1}
	for _, structure := range []Structure{StructureFFA, StructureCoOp, StructureAutoBalanced} {
		config := Config{Structure: structure}
		if structure == StructureAutoBalanced {
			config.AutoTeamCount = 2
		}
		if _, err := ResolveAssignments(config, []string{"player-1"}, requested); err == nil {
			t.Errorf("ResolveAssignments(%q) accepted requested assignments", structure)
		}
	}
}

func TestResolveAssignmentsValidatesBeforeDispatch(t *testing.T) {
	config := Config{Structure: Structure("unknown")}
	if _, err := ResolveAssignments(config, []string{"", ""}, nil); err == nil {
		t.Fatal("ResolveAssignments did not reject invalid config before roster resolution")
	}
}

func TestResolveAssignmentsIsDeterministicAndDoesNotMutateInputs(t *testing.T) {
	config := Config{Structure: StructureAutoBalanced, AutoTeamCount: 3}
	participants := []string{"player-3", "player-1", "player-2"}
	requested := Assignments{}
	first, err := ResolveAssignments(config, participants, requested)
	if err != nil {
		t.Fatalf("ResolveAssignments() error = %v", err)
	}
	second, err := ResolveAssignments(config, participants, requested)
	if err != nil {
		t.Fatalf("ResolveAssignments() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("identical inputs produced different outputs: %v vs %v", first, second)
	}
	if !reflect.DeepEqual(participants, []string{"player-3", "player-1", "player-2"}) {
		t.Fatal("ResolveAssignments mutated participant IDs")
	}
	if len(requested) != 0 {
		t.Fatal("ResolveAssignments mutated requested assignments")
	}
}
