package teams

import (
	"reflect"
	"testing"
)

func TestResolveCustomAllowsUnevenAssignmentsAndAllSlots(t *testing.T) {
	participants := []string{"player-2", "player-1", "player-3"}
	requested := Assignments{
		"player-1": Team8,
		"player-2": Team1,
		"player-3": Team1,
	}
	got, err := ResolveCustom(participants, requested)
	if err != nil {
		t.Fatalf("ResolveCustom() error = %v", err)
	}
	if !reflect.DeepEqual(got, requested) {
		t.Fatalf("ResolveCustom() = %v, want %v", got, requested)
	}

	all := make(Assignments, 8)
	for index, teamID := range OrderedIDs() {
		all[participantID(index)] = teamID
	}
	if _, err := ResolveCustom(participantIDs(8), all); err != nil {
		t.Fatalf("ResolveCustom() rejected all eight slots: %v", err)
	}
}

func TestResolveCustomDoesNotMutateRequested(t *testing.T) {
	requested := Assignments{"player-2": Team2, "player-1": Team1}
	original := Assignments{"player-2": Team2, "player-1": Team1}
	if _, err := ResolveCustom([]string{"player-2", "player-1"}, requested); err != nil {
		t.Fatalf("ResolveCustom() error = %v", err)
	}
	if !reflect.DeepEqual(requested, original) {
		t.Fatal("ResolveCustom mutated requested assignments")
	}
}

func TestResolveCustomRejectsInvalidAssignments(t *testing.T) {
	tests := []struct {
		name      string
		roster    []string
		requested Assignments
	}{
		{name: "missing", roster: []string{"player-1", "player-2"}, requested: Assignments{"player-1": Team1}},
		{name: "unknown participant", roster: []string{"player-1"}, requested: Assignments{"player-1": Team1, "player-2": Team2}},
		{name: "no team", roster: []string{"player-1"}, requested: Assignments{"player-1": NoTeam}},
		{name: "invalid team", roster: []string{"player-1"}, requested: Assignments{"player-1": ID("team_9")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ResolveCustom(test.roster, test.requested); err == nil {
				t.Fatal("ResolveCustom returned nil error")
			}
		})
	}
}

func participantID(index int) string {
	return "player-" + string(rune('1'+index))
}

func participantIDs(count int) []string {
	participants := make([]string, count)
	for index := range participants {
		participants[index] = participantID(index)
	}
	return participants
}
